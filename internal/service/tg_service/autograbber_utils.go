package tg_service

import (
	"fmt"
	"myapp/internal/entity"
	"myapp/internal/models"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// метод заменяет ссылку на канал и пост такого вида https://t.me/c/1949679854/4333, под конкретного vampBota
func (srv *TgService) ChangeLinkReferredToPost(originalLink string, vampBot entity.Bot) (string, error) {
	urlArr := strings.Split(originalLink, "/")
	for i, v := range urlArr {
		if len(urlArr) < 4 {
			break
		}
		if v == "t.me" && urlArr[i+1] == "c" {
			chId := urlArr[i+2]
			postId := urlArr[i+3]
			logMes := fmt.Sprintf("ChangeLinkReferredToPost: это ссылка на канал %s и пост %s", chId, postId)
			srv.l.Info(logMes)

			refToDonorChPostId, err := strconv.Atoi(postId)
			if err != nil {
				return "", fmt.Errorf("ChangeLinkToPost Atoi err: %v", err)
			}
			currPost, err := srv.db.GetPostByDonorIdAndChId(refToDonorChPostId, vampBot.ChId)
			if err != nil {
				return "", fmt.Errorf("ChangeLinkToPost GetPostByDonorIdAndChId err: %v", err)
			}
			if vampBot.ChId < 0 {
				urlArr[i+2] = strconv.Itoa(-vampBot.ChId)
			} else {
				urlArr[i+2] = strconv.Itoa(vampBot.ChId)
			}
			if urlArr[i+2][0] == '1' && urlArr[i+2][1] == '0' && urlArr[i+2][2] == '0' {
				urlArr[i+2] = urlArr[i+2][3:]
			}
			urlArr[i+3] = strconv.Itoa(currPost.PostId)

			newLink := strings.Join(urlArr, "/")
			return newLink, nil
		}
	}
	// https://t.me/lichka
	if strings.HasPrefix(originalLink, "http://t.me/lichka") || strings.HasPrefix(originalLink, "t.me/lichka") || strings.HasPrefix(originalLink, "https://t.me/lichka") {
		lichka := vampBot.Lichka
		if lichka != "" {
			newLink := fmt.Sprintf("https://t.me/%s", srv.DelAt(lichka))
			return newLink, nil
		}
	}
	return "", nil
}

// метод заменяет fake-link на нужную группу-ссылку vampBota
// и вырезает все ссылки и Entities если группа-ссылка - cut-link
func (srv *TgService) PrepareEntities(entities []models.MessageEntity, messText string, vampBot entity.Bot) ([]models.MessageEntity, string, error) {
	srv.l.Info("PrepareEntities", zap.Any("vampBot", vampBot))
	cutEntities := false
	for i, v := range entities {
		// если fake-link
		if strings.HasPrefix(v.Url, "http://fake-link") || strings.HasPrefix(v.Url, "fake-link") || strings.HasPrefix(v.Url, "https://fake-link") {
			groupLink, err := srv.db.GetGroupLinkById(vampBot.GroupLinkId)
			if err != nil {
				return nil, messText, err
			}
			srv.l.Info("PrepareEntities:", zap.Any("vampBot", vampBot), zap.Any("groupLink", groupLink))
			if groupLink.Link == "" {
				continue
			}
			// если cut-link
			if strings.HasPrefix(groupLink.Link, "http://cut-link") || strings.HasPrefix(groupLink.Link, "cut-link") || strings.HasPrefix(groupLink.Link, "https://cut-link") {
				messText = strings.Replace(messText, "Переходим по ссылке - ССЫЛКА", "", -1)
				messText = strings.Replace(messText, "👉 РЕГИСТРАЦИЯ ТУТ 👈", "", -1)
				messText = strings.Replace(messText, "🔖 Написать мне 🔖", "", -1)
				cutEntities = true
				break
			}
			refLink := groupLink.Link
			if srv.Cfg.IsPersonalLinks == 1 {
				if vampBot.PersonalLink != "" {
					refLink = vampBot.PersonalLink
				}
			}
			entities[i].Url = refLink
			continue
		}
		// если Tg ссылка
		newUrl, err := srv.ChangeLinkReferredToPost(v.Url, vampBot)
		if err != nil {
			return nil, messText, fmt.Errorf("PrepareEntities ChangeLinkReferredToPost err: %v", err)
		}
		if newUrl != "" {
			entities[i].Url = newUrl
		}
		srv.l.Info("PrepareEntities", zap.Any("newUrl", newUrl), zap.Any("vampBot", vampBot))
	}
	lichka := srv.AddAt(vampBot.Lichka)
	srv.l.Info("PrepareEntities Replace 1 @lichka", zap.Any("lichka", lichka), zap.Any("old messText", messText), zap.Any("vampBot", vampBot))
	if lichka != "" {
		messText = strings.Replace(messText, "@lichka", lichka, -1)
	}
	srv.l.Info("PrepareEntities Replace 2 @lichka", zap.Any("lichka", lichka), zap.Any("new messText", messText), zap.Any("vampBot", vampBot))
	if !cutEntities {
		return entities, messText, nil
	}
	return nil, messText, nil
}

func (srv *TgService) PrepareReplyMarkup(entities models.InlineKeyboardMarkup, vampBot entity.Bot) (models.InlineKeyboardMarkup, error) {
	for i, v := range entities.InlineKeyboard {
		for ii, vv := range v {
			if vv.Url == nil {
				continue
			}
			// если fake-link
			if strings.HasPrefix(*vv.Url, "http://fake-link") || strings.HasPrefix(*vv.Url, "fake-link") || strings.HasPrefix(*vv.Url, "https://fake-link") {
				groupLink, err := srv.db.GetGroupLinkById(vampBot.GroupLinkId)
				if err != nil {
					return models.InlineKeyboardMarkup{}, err
				}
				srv.l.Info("PrepareEntities:", zap.Any("vampBot", vampBot), zap.Any("groupLink", groupLink))
				if groupLink.Link == "" {
					continue
				}
				entities.InlineKeyboard[i][ii].Url = &groupLink.Link
				continue
			}
			// если Tg ссылка
			newUrl, err := srv.ChangeLinkReferredToPost(*vv.Url, vampBot)
			if err != nil {
				return models.InlineKeyboardMarkup{}, fmt.Errorf("PrepareReplyMarkup ChangeLinkReferredToPost err: %v", err)
			}
			if newUrl != "" {
				entities.InlineKeyboard[i][ii].Url = &newUrl
			}
		}
	}
	return entities, nil
}

func (srv *TgService) GetPostAndChFromLink(link string) (string, string, error) {
	urlArr := strings.Split(link, "/")
	if len(urlArr) != 6 {
		return "", "", fmt.Errorf("GetPostAndChFromLink err: не правилная ссылка %s", link)
	}
	for i, v := range urlArr {
		if v == "t.me" && urlArr[i+1] == "c" {
			chId := urlArr[i+2]
			postId := urlArr[i+3]
			logMes := fmt.Sprintf("GetPostAndChFromLink: это ссылка на канал %s и пост %s", chId, postId)
			srv.l.Info(logMes)
			return chId, postId, nil
		}
	}
	return "", "", nil
}