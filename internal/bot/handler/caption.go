package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/and3rson/telemux/v2"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slices"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"markoslav/internal/bot/template"
	"markoslav/internal/dto"
	"markoslav/internal/model"
	"markoslav/internal/usecase"
	"markoslav/pkg/apperror"
	"markoslav/pkg/filter"
	"math/rand"
	"net/http"
	"time"
)

const (
	RandomCaptionCommand = "марк"
	HelpMessageText      = `
Вы можете управлять мной, посылая эти команды (только в приватном диалоге)

/suggest - предложить новую подпись
/approve - просмотр предложенных подписей (только для администрации)
/cancel - отменить текущую команду
`
	UnknownErrorMessageText = "Произошла непредвиденная ошибка."
)

type CaptionHandler struct {
	api            *tgbotapi.BotAPI
	captionUsecase usecase.CaptionUsecase
	adminList      []int64
}

func NewCaptionHandler(api *tgbotapi.BotAPI, captionUsecase usecase.CaptionUsecase, adminList []int64) *CaptionHandler {
	return &CaptionHandler{api: api, captionUsecase: captionUsecase, adminList: adminList}
}

func (handler *CaptionHandler) Register(mux *telemux.Mux) {
	mux.AddHandler(
		telemux.NewCommandHandler(
			"start",
			telemux.Any(),
			func(update *telemux.Update) {
				handler.api.Send(tgbotapi.NewMessage(update.EffectiveChat().ID, HelpMessageText))
			},
		),
		telemux.NewCommandHandler(
			"help",
			telemux.Any(),
			func(update *telemux.Update) {
				handler.api.Send(tgbotapi.NewMessage(update.EffectiveChat().ID, HelpMessageText))
			},
		),
		telemux.NewConversationHandler(
			"approve_captions",
			telemux.NewLocalPersistence(),
			telemux.StateMap{
				"": {
					telemux.NewCommandHandler(
						"approve",
						telemux.And(telemux.IsPrivate(), isAdmin(handler.adminList)),
						func(update *telemux.Update) {
							options := filter.NewOptions().
								Add("approved", false, filter.OperatorEq)

							chat := update.EffectiveChat()

							captions, err := handler.captionUsecase.Select(context.TODO(), 25, 0, options)
							if err != nil {
								handler.api.Send(tgbotapi.NewMessage(chat.ID, UnknownErrorMessageText))

								log.Println(err)
								return
							}

							if len(captions) == 0 {
								handler.api.Send(tgbotapi.NewMessage(chat.ID, "Не удалось найти подписи на одобрение."))
								return
							}

							update.PersistenceContext.PutDataValue("captions", captions)
							update.PersistenceContext.PutDataValue("reviewed_caption_index", 0)

							text, markup := ApprovingCaptionsMessageText(update)
							message := tgbotapi.NewMessage(chat.ID, text)
							message.ReplyMarkup = markup

							if _, err = handler.api.Send(message); err != nil {
								log.Println(err)
								return
							}

							update.PersistenceContext.SetState("approving_captions")
						},
					),
				},
				"approving_captions": {
					telemux.NewCallbackQueryHandler(
						"approve_caption",
						telemux.Any(),
						func(update *telemux.Update) {
							data := update.PersistenceContext.GetData()
							captions := data["captions"].([]model.Caption)
							reviewedCaptionIndex := data["reviewed_caption_index"].(int)

							caption := captions[reviewedCaptionIndex]
							err := handler.captionUsecase.Approve(context.TODO(), caption.ID)
							if err != nil {
								log.Println(err)
								return
							}

							update.PersistenceContext.PutDataValue("reviewed_caption_index", reviewedCaptionIndex+1)

							text, markup := ApprovingCaptionsMessageText(update)

							message := update.EffectiveMessage()
							edit := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, text)
							if markup != nil {
								edit.ReplyMarkup = markup
							}

							if _, err := handler.api.Send(edit); err != nil {
								log.Println(err)
							}
						},
					),
					telemux.NewCallbackQueryHandler(
						"reject_caption",
						telemux.Any(),
						func(update *telemux.Update) {
							data := update.PersistenceContext.GetData()
							captions := data["captions"].([]model.Caption)
							reviewedCaptionIndex := data["reviewed_caption_index"].(int)

							caption := captions[reviewedCaptionIndex]
							err := handler.captionUsecase.Reject(context.TODO(), caption.ID)
							if err != nil {
								log.Println(err)
								return
							}

							update.PersistenceContext.PutDataValue("reviewed_caption_index", reviewedCaptionIndex+1)

							text, markup := ApprovingCaptionsMessageText(update)

							message := update.EffectiveMessage()
							edit := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, text)
							if markup != nil {
								edit.ReplyMarkup = markup
							}

							if _, err := handler.api.Send(edit); err != nil {
								log.Println(err)
							}
						},
					),
				},
			},
			[]*telemux.Handler{
				telemux.NewCallbackQueryHandler(
					"cancel",
					telemux.Any(),
					func(update *telemux.Update) {
						message := update.EffectiveMessage()

						reply := tgbotapi.NewEditMessageText(
							message.Chat.ID,
							message.MessageID,
							"Команда /approve была успешно отменена.",
						)

						if _, err := handler.api.Send(reply); err != nil {
							log.Println(err)
							return
						}

						update.PersistenceContext.ClearData()
						update.PersistenceContext.SetState("")
					},
				),
			},
		),
		telemux.NewConversationHandler(
			"suggest_caption",
			telemux.NewLocalPersistence(),
			telemux.StateMap{
				"": {
					telemux.NewCommandHandler(
						"suggest",
						telemux.IsPrivate(),
						func(update *telemux.Update) {
							message := tgbotapi.NewMessage(
								update.Message.Chat.ID,
								"Отправьте в чат подпись, которую вы хотите предложить.",
							)

							if _, err := handler.api.Send(message); err != nil {
								log.Println(err)
							}

							update.PersistenceContext.SetState("enter_caption")
						},
					),
				},
				"enter_caption": {
					telemux.NewMessageHandler(
						telemux.HasText(),
						func(update *telemux.Update) {
							clearState := true

							message := update.EffectiveMessage()

							reply := tgbotapi.NewMessage(message.Chat.ID, "Подпись была успешно отправлена на подтверждение.")

							_, err := handler.captionUsecase.Create(context.TODO(), dto.CreateCaption{
								Text:     message.Text,
								AuthorID: message.From.ID,
							})
							if err != nil {
								detail := UnknownErrorMessageText

								if _, ok := apperror.Is(err, apperror.Internal); ok {
									log.Printf("enter caption: %s\n", err)
								} else if _, ok = apperror.Is(err, apperror.AlreadyExists); ok {
									detail = "Такая подпись уже существует. Попробуйте что-нибудь другое."
									clearState = false
								}

								reply.Text = fmt.Sprintf("Не удалось отправить подпись. %s", detail)
							}

							if _, err := handler.api.Send(reply); err != nil {
								log.Println(err)
								return
							}

							if clearState {
								update.PersistenceContext.ClearData()
								update.PersistenceContext.SetState("")
							}
						},
					),
				},
			},
			[]*telemux.Handler{
				telemux.NewCommandHandler(
					"cancel",
					telemux.Any(),
					func(update *telemux.Update) {
						message := update.EffectiveMessage()

						reply := tgbotapi.NewMessage(
							message.Chat.ID,
							"Команда /suggest была успешно отменена.",
						)

						if _, err := handler.api.Send(reply); err != nil {
							log.Println(err)
							return
						}

						update.PersistenceContext.ClearData()
						update.PersistenceContext.SetState("")
					},
				),
			},
		),
		telemux.NewMessageHandler(
			func(update *telemux.Update) bool {
				message := update.Message
				text := message.Text

				if message.Caption != "" {
					text = message.Caption
				}

				isReply := false
				if message.ReplyToMessage != nil {
					message = message.ReplyToMessage
					isReply = true
				}

				if text == RandomCaptionCommand || ((rand.Intn(100) < 50) && isReply == false) && len(message.Photo) > 0 {
					update.Context["photo"] = message.Photo[len(message.Photo)-1]

					return true
				}

				return false
			},
			func(update *telemux.Update) {
				photo := update.Context["photo"].(tgbotapi.PhotoSize)

				fileURL, err := handler.api.GetFileDirectURL(photo.FileID)
				if err != nil {
					log.Printf("get file direct url: %s", err)
					return
				}

				var img image.Image
				img, err = fetchImage(fileURL)
				if err != nil {
					log.Printf("fetch image: %s", err)
					return
				}

				img, err = handler.captionUsecase.DrawRandom(context.TODO(), img)
				if err != nil {
					log.Printf("draw random caption: %s", err)
					return
				}

				buffer := new(bytes.Buffer)
				if err = png.Encode(buffer, img); err != nil {
					log.Printf("encode image: %s", err)
					return
				}

				photoConfig := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileBytes{
					Name:  "picture",
					Bytes: buffer.Bytes(),
				})
				photoConfig.ReplyToMessageID = update.Message.MessageID

				if _, err = handler.api.Send(photoConfig); err != nil {
					log.Printf("failed to send: %s", err)
					return
				}
			},
		),
	)
}

func fetchImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", response.StatusCode)
	}

	var img image.Image
	img, _, err = image.Decode(response.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func isAdmin(adminList []int64) telemux.FilterFunc {
	return func(update *telemux.Update) bool {
		return slices.Contains(adminList, update.EffectiveUser().ID)
	}
}

func ApprovingCaptionsMessageText(update *telemux.Update) (string, *tgbotapi.InlineKeyboardMarkup) {
	data := update.PersistenceContext.GetData()
	captions := data["captions"].([]model.Caption)
	reviewedCaptionIndex := data["reviewed_caption_index"].(int)

	if reviewedCaptionIndex >= len(captions) {
		update.PersistenceContext.ClearData()
		update.PersistenceContext.SetState("")

		return "Подписи на одобрение закончились", nil
	}

	caption := captions[reviewedCaptionIndex]

	buffer := new(bytes.Buffer)
	err := template.ApproveCaptions.Execute(buffer, map[string]any{
		"reviewed_count":             len(captions) - reviewedCaptionIndex,
		"total_disapproved_remained": len(captions),
		"text":                       caption.Text,
		"author_id":                  caption.AuthorID,
		"created_at":                 caption.CreatedAt.Format(time.RFC3339),
	})
	if err != nil {
		update.PersistenceContext.ClearData()
		update.PersistenceContext.SetState("")

		return "Произошла непредвиденная ошибка.", nil
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Одобрить", "approve_caption"),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", "reject_caption"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
		),
	)

	return buffer.String(), &keyboard
}
