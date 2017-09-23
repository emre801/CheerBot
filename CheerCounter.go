package main

import ( 
	"encoding/json"
	"io/ioutil"
    "os"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"net"
	"net/textproto"
	"bufio"
)

type Page struct {
    Channel    	string	`json:"channel"`
    Username 	string 	`json:"username"`
	Oauth   	string 	`json:"oauth"`
	CommandA 	string 	`json:"commandA"`
	CommandB 	string 	`json:"commandB"`
	KeyWorld 	string 	`json:"keyWord"`
}
type Score struct {
    A   int	`json:"a"`
    B 	int `json:"b"`
}

func (p Page) toString() string {
    return toJson(p)
}
func toString(s Score) string {
    return toJson(s)
}
func toJson(p interface{}) string {
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
	s := " " + s2 + " "
	r := regexp.MustCompile("\\s("+page.KeyWorld+")(.*?[\\d]+)\\s")
	f := r.FindStringSubmatch(s)
	if(len(f) == 3){
		m := f[2]
		i,_ := strconv.ParseInt(m, 0, 64)
		if(strings.Contains(s,"#" + page.CommandA)){
			return int(i);
		} else if(strings.Contains(s,"#" + page.CommandB)){
			return int(i * -1);
		}
	} 
	return 0
}

func goBotGo(page Page, score Score) {
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
	purple := score.A
	pink := score.B

	for {
        msg, err := tp.ReadLine()
        if err != nil {
            panic(err)    
		}
		fmt.Println(msg)
		
		splitMessage := strings.Split(strings.ToLower(msg), ":")
		if(len(splitMessage) >= 3) {
			cheer := countCheers(strings.Split(msg, ":")[2], page)
			//fmt.Println(strings.Split(msg, ":")[2])
			//fmt.Println(cheer)
			if(cheer > 0) {
				pink += cheer
			} else if (cheer < 0) {
				purple +=  cheer * -1
			}

			if cheer != 0 {
				pinkString := strconv.Itoa(pink)
				purpleString := strconv.Itoa(purple)
				score.B = purple
				score.A = pink
				json := "[" + toString(score) + "]"
				ioutil.WriteFile("score.json", []byte(json), 0644)
				//fmt.Println(" Pink : " + pinkString +  " Purple : " + purpleString + "\r\n")
				conn.Write([]byte("PRIVMSG #"+page.Channel +  " :" + "Pink: " + pinkString +  " Purple: " + purpleString+ "\r\n"))
				continue
			} 
		}
    }
}

func main() {
	pages := getPages()
	score := getScore()
	goBotGo(pages[0], score[0])
}