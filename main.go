package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	token              = os.Getenv("TOKEN")
	apiEndpoint        = "https://api.telegram.org/bot"
	requestURL         = apiEndpoint + token
	groupNumberPattern = "(ЕІ-81)|(П-81)"
)

var tableEmoji = map[int]string{
	0: "👩‍🎓",
	1: "⏰",
	2: "📚",
	3: "🎉",
	4: "🚪",
}

var sheduleTableEmoji = map[int]string{
	0: "📚",
	1: "👩‍🎓",
	2: "🚪",
	3: "⏰",
	4: "📆",
}

// ASCIIEmoji used instead of empty substitutions
const ASCIIEmoji = "(•ω•)⊃──☆ﾟ.･❁｡ﾟ✧"

func main() {
	if len(token) == 0 {
		log.Fatal("Environment variable `TOKEN` must be defined")
	}
	go watchForSubstitutionsUpdate()
	updates := make(chan Update)
	go startPolling(updates)
	for {
		select {
		case upd := <-updates:
			processUpdate(upd)
		}
	}
}

func startPolling(c chan Update) {
	offset := 0
	for {
		requestParameters := url.Values{
			"offset":  {fmt.Sprintf("%v", offset)},
			"timeout": {"1"},
		}
		data, err := MakeTgapiRequest("getUpdates", requestParameters)
		if err != nil {
			log.Println(err)
		}
		var updates Updates
		err = json.Unmarshal(data, &updates)
		if err != nil {
			log.Println(err)
		}
		if updates.Ok {
			if r := updates.Result; offset == 0 && len(r) > 0 {
				offset = r[len(r)-1].UpdateID + 1
			} else {
				for idx, result := range updates.Result {
					c <- result
					if idx == len(updates.Result)-1 {
						offset = result.UpdateID + 1
					}
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func processUpdate(u Update) {
	_, err := strconv.Atoi(u.CallbackQuery.ID)
	if err == nil {
		HandleCallbackQuery(u.CallbackQuery)
	}
	msgText := u.Message.Text
	for pattern, handler := range MessageHandlers {
		matched, err := regexp.MatchString(pattern, msgText)
		if err != nil {
			continue
		}
		if matched {
			handler(u.Message)
			return
		}
	}
}

// SendMessage sends message with provided text to chat that has given id
func SendMessage(text string, chatID int) {
	p := url.Values{
		"chat_id": {fmt.Sprintf("%v", chatID)},
		"text":    {text},
	}
	MakeTgapiRequest("sendMessage", p)
}

// MakeTgapiRequest does a request by given method name
func MakeTgapiRequest(methodName string, parameters url.Values) ([]byte, error) {
	res, err := http.PostForm(requestURL+"/"+methodName, parameters)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetSubstitutions fetches html page and returns it parsed into message text
func GetSubstitutions() (string, error) {
	doc, err := getSubstitutionsDocument()
	if err != nil {
		return "", err
	}
	message, err := parseDocument(doc)
	if err != nil {
		return "", err
	}

	return message, nil
}

func getSubstitutionsDocument() (*goquery.Document, error) {
	res, err := http.Get("http://ki.sumdu.edu.ua/zamen/mes_inst.html")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Parse document structure into message text
func parseDocument(doc *goquery.Document) (string, error) {
	var headerText, bodyText, classroomSubstitutions string
	var err error
	headerText, err = parseDocumentHeader(doc)
	if err != nil {
		return "", err
	}
	bodyText, err = parseLessonSubstitutions(doc)
	if err != nil {
		return "", err
	}
	classroomSubstitutions, err = parseClassroomSubstitutions(doc)
	if err != nil {
		return "", err
	}
	link := "[Переглянути на сайті 🦄](http://ki.sumdu.edu.ua/zamen/mes_inst.html)"
	message := fmt.Sprintf("%v\n\n%v\n\n%v%v", headerText, bodyText, classroomSubstitutions, link)
	return message, nil
}

func parseDocumentHeader(doc *goquery.Document) (string, error) {
	header := doc.Find("body>div>p").First()
	html, err := header.Html()
	if err != nil {
		return "", err
	}
	html = regexp.MustCompile(`<br/>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`\n{2,}`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`^\s+`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</*\w+>`).ReplaceAllString(html, "")
	headerText := strings.TrimSpace(html)
	return headerText, nil
}

func parseLessonSubstitutions(doc *goquery.Document) (string, error) {
	var header, substitutions string
	headerTitles := map[int]string{
		0: "Група",
		1: "Пара",
		2: "Предмет",
		3: "Заміна",
		4: "Аудиторія",
	}
	table := doc.Find("table").First()
	table.Find("tr").Each(func(idx int, sel *goquery.Selection) {
		if idx == 0 {
			sel.Children().Each(func(i int, s *goquery.Selection) {
				header += fmt.Sprintf("%v %v \n", tableEmoji[i], headerTitles[i])
			})
		} else {
			matched := isMatchedGroupNumberForTableRow(sel)
			if !matched {
				return
			}
			sel.Children().Each(func(i int, s *goquery.Selection) {
				t := s.Text()
				if len(t) < 1 {
					return
				}
				substitutions += fmt.Sprintf("%v %v \n", tableEmoji[i], t)
			})
		}
		substitutions += "\n"
	})
	if s := strings.TrimSpace(substitutions); len(s) == 0 {
		substitutions = fmt.Sprintf("%v\nНемає замін 🙂", ASCIIEmoji)
		return substitutions, nil
	}
	substitutions = header + substitutions
	substitutions = regexp.MustCompile(`\n{2,}$`).ReplaceAllString(substitutions, "\n")
	substitutions = strings.TrimSpace(substitutions)
	return substitutions, nil
}

func parseClassroomSubstitutions(doc *goquery.Document) (string, error) {
	var classroomSubstitutions string
	emojis := map[int]string{
		0: tableEmoji[0],
		1: tableEmoji[1],
		2: tableEmoji[4],
	}
	table := doc.Find("table").Last()
	table.Find("tr").Each(func(idx int, sel *goquery.Selection) {
		matched := isMatchedGroupNumberForTableRow(sel)
		if !matched {
			return
		}
		sel.Children().Each(func(i int, td *goquery.Selection) {
			text := td.Text()
			emoji := emojis[i]
			classroomSubstitutions += fmt.Sprintf("%v %v\n", emoji, text)
		})
		classroomSubstitutions += "\n"
	})
	if s := strings.TrimSpace(classroomSubstitutions); len(s) == 0 {
		return "", nil
	}
	classroomSubstitutions = "Заміна аудиторій 🎈\n\n" + classroomSubstitutions
	classroomSubstitutions = strings.TrimSpace(classroomSubstitutions)
	classroomSubstitutions += "\n\n"
	return classroomSubstitutions, nil
}

func isMatchedGroupNumberForTableRow(row *goquery.Selection) bool {
	group := row.Children().First().Text()
	matched, err := regexp.Match(groupNumberPattern, []byte(group))
	if err != nil {
		return false
	}
	return matched
}

func watchForSubstitutionsUpdate() {
	var data []byte
	c := time.Tick(5 * time.Minute)
	for range c {
		targetChatIDs := os.Getenv("CHAT_ID")
		if len(targetChatIDs) == 0 {
			continue
		}
		res, err := http.Get("http://ki.sumdu.edu.ua/zamen/mes_inst.html")
		if err != nil {
			continue
		}
		defer res.Body.Close()
		response, err := ioutil.ReadAll(res.Body)
		if err != nil {
			continue
		}
		if len(data) > 0 && reflect.DeepEqual(data, response) == false {
			targetChats := strings.Split(targetChatIDs, ",")
			for _, targetChatID := range targetChats {
				chatID, err := strconv.Atoi(targetChatID)
				if err != nil {
					continue
				}
				SendSubstitutions(chatID)
				time.Sleep(time.Second)
			}
		}
		data = response
	}
}

// RespondLessonsSheduleCallbackQuery asks user to select group
func RespondLessonsSheduleCallbackQuery(cq CallbackQuery) {
	groupName := ""
	p := url.Values{}

	switch cq.Data {
	case "group_ei81":
		groupName = "ЕІ-81"
	case "group_p81":
		groupName = "П-81"
	}

	p.Set("nomer_grup", groupName)

	res, err := http.PostForm("https://ki.sumdu.edu.ua/?p=612", p)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}

	message := fmt.Sprintf("Розклад занять для групи %v 🥳\n\n", groupName)

	doc.Find(".post table tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td").Each(func(n int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			emoji := sheduleTableEmoji[n]
			message += fmt.Sprintf("%v %v\n", emoji, text)
		})
		if i == 0 {
			message = fmt.Sprintf("*%v*", message)
		}
		message += "\n"
	})

	v := url.Values{
		"chat_id":    {fmt.Sprintf("%v", cq.Message.Chat.ID)},
		"message_id": {fmt.Sprintf("%v", cq.Message.ID)},
		"text":       {message},
		"parse_mode": {"Markdown"},
	}
	MakeTgapiRequest("editMessageText", v)
}
