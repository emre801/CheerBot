package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

//Page are settings
type Page struct {
	Channel  string   `json:"channel"`
	Username string   `json:"username"`
	Oauth    string   `json:"oauth"`
	CommandA string   `json:"commandA"`
	CommandB string   `json:"commandB"`
	KeyWorld []string `json:"keyWord"`
}

//Score that is saved
type Score struct {
	A int `json:"a"`
	B int `json:"b"`
}

var i = 0
var j = 0

func (p Page) toString() string {
	return toJSON(p)
}

func toString(s Score) string {
	return toJSON(s)
}

func toJSON(p interface{}) string {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return string(bytes)
}

func getPages() []Page {
	raw, err := ioutil.ReadFile("./settings.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []Page
	json.Unmarshal(raw, &c)
	return c
}

func getScore() []Score {
	raw, err := ioutil.ReadFile("./score.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []Score
	json.Unmarshal(raw, &c)
	return c
}

func countCheers(s2 string, page Page) int {
	result := 0
	for _, keyword := range page.KeyWorld {
		result += countCheersSingle(s2, keyword, page.CommandA, page.CommandB)
	}
	return result
}

func countCheersSingle(s2 string, keyword string, commandA string, commandB string) int {

	s := strings.Replace(" "+s2+" ", " ", "  ", -1)
	r := regexp.MustCompile("\\s(" + keyword + ")(.*?[\\d]+)\\s")
	ff := r.FindAllStringSubmatch(s, 40)
	result := 0
	for _, f := range ff {
		fmt.Println(f)
		// element is the element from someSlice for where we are
		if len(f) == 3 {
			m := f[2]
			i, _ := strconv.ParseInt(m, 0, 64)
			if strings.Contains(s, "#"+commandA) {
				result += int(i)
			} else if strings.Contains(s, "#"+commandB) {
				result += int(i * -1)
			}
		}
	}

	return result
}

func goBotGo(page Page, score Score, win *pixelgl.Window) {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("PASS " + page.Oauth + "\r\n"))
	conn.Write([]byte("NICK " + page.Username + "\r\n"))
	conn.Write([]byte("JOIN " + "#" + page.Channel + "\r\n"))
	defer conn.Close()

	tp := textproto.NewReader(bufio.NewReader(conn))
	fmt.Println(strconv.Itoa(score.A))
	fmt.Println(strconv.Itoa(score.B))
	scoreA := score.A
	scoreB := score.B
	i = pink
	j = scoreA

	for !win.Closed() {
		msg, err := tp.ReadLine()
		if err != nil {
			panic(err)
		}
		fmt.Println(msg)
		msgParts := strings.Split(msg, " ")
		if msgParts[0] == "PING" {
			conn.Write([]byte("PONG " + msgParts[1]))
			continue
		}

		splitMessage := strings.Split(strings.ToLower(msg), ":")
		if len(splitMessage) >= 3 {
			if strings.Split(msg, ":")[2] == "!score" {
				outputScore(scoreB, scoreA, conn, score, page)
				continue
			}

			cheer := countCheers(strings.Split(msg, ":")[2], page)
			fmt.Println(strings.Split(msg, ":")[2])
			fmt.Println(cheer)
			if cheer > 0 {
				scoreB += cheer
			} else if cheer < 0 {
				scoreA += cheer * -1
			}

			if cheer != 0 {
				outputScore(scoreB, scoreA, conn, score, page)
				continue
			}
		}
	}
}

func outputScore(scoreB int, scoreA int, conn net.Conn, score Score, page Page) {
	scoreAString := strconv.Itoa(scoreA)
	scoreBString := strconv.Itoa(scoreA)
	score.B = scoreA
	score.A = scoreA
	json := "[" + toString(score) + "]"
	ioutil.WriteFile("score.json", []byte(json), 0644)
	i = scoreA
	j = scoreA
	fmt.Println(" " + page.CommandA + "  : " + scoreAString + " " + page.CommandB + " : " + scoreBString + "\r\n")
	conn.Write([]byte("PRIVMSG #" + page.Channel + " :" + "Pink: " + scoreAString + " Purple: " + scoreBString + "\r\n"))
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Cheer Bot",
		Bounds: pixel.R(0, 0, 600, 175),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Black)

	pages := getPages()
	score := getScore()

	go goBotGo(pages[0], score[0], win)
	runWindow(win)

}

func runWindow(win *pixelgl.Window) {

	for !win.Closed() {
		imd := imdraw.New(nil)
		basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
		pinkTXT := text.New(pixel.V(10, 150), basicAtlas)
		fmt.Fprintln(pinkTXT, "PINK")
		purpleTXT := text.New(pixel.V(400, 150), basicAtlas)
		fmt.Fprintln(purpleTXT, "PURPLE")
		if i > 0 || j > 0 {
			totalVotes := float64(i) + float64(j)
			iPercent := float64(i) / totalVotes * 600

			r := float64(iPercent)
			//purple
			imd.Color = pixel.RGB(85.0/256.0, 26.0/256.0, 139.0/256.0)
			imd.Push(pixel.V(0, 0))
			imd.Push(pixel.V(0, 100))
			imd.Push(pixel.V(600, 100))
			imd.Push(pixel.V(600, 0))
			imd.Polygon(0)

			//pink
			imd.Color = pixel.RGB(255.0/256.0, 105.0/256.0, 180.0/256)
			imd.Push(pixel.V(0, 0))
			imd.Push(pixel.V(0, 100))
			imd.Push(pixel.V(r, 100))
			imd.Push(pixel.V(r, 0))
			imd.Polygon(0)
			purpleTXT.Clear()
			pinkTXT.Clear()
			fmt.Fprintln(purpleTXT, strconv.Itoa(j)+" : PURPLE")
			fmt.Fprintln(pinkTXT, "PINK : "+strconv.Itoa(i))
		}
		win.Clear(colornames.Black)
		imd.Draw(win)
		purpleTXT.Draw(win, pixel.IM.Scaled(purpleTXT.Orig, 2))
		pinkTXT.Draw(win, pixel.IM.Scaled(pinkTXT.Orig, 2))
		win.Update()
		time.Sleep(time.Millisecond * 16) //Sleep for a frame
	}
}
func runImage() {
	pixelgl.Run(run)
}

func main() {
	runImage()
}
