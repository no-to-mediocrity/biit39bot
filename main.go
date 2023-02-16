package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {

	log.Println("Setting up the environment...")
	//setting up working dir
	cwd, err := os.Getwd()
	checkerr(err, "Error while getting current working directory: ")

	filepathnew := filepath.ToSlash(path.Join(cwd, "schedule_new.pdf"))
	filepathold := filepath.ToSlash(path.Join(cwd, "schedule.pdf"))

	file, err := os.Open(filepathold)
	if err != nil {
		// No previous version, save the current version
		log.Println("No schedule files found, getting the file from biit39.ru...")
		file, err = os.Create(filepathold)
		checkerr(err, "Error while creating the schedule file: ")
		schedule, err := getSchedule()
		defer schedule.Body.Close()
		checkerr(err)
		err = saveFile(filepathold, schedule)
		if err != nil {
			log.Println("Error while retrieving schedule: ", err)
		} else {
			log.Println("Got the schedule from biit39.ru for the first time")
		}
	}
	file.Close()

	log.Println("Logging in to telegram...")
	token := `5893257540:AAG4UJUCKCuxwtFP6IpRYEgurC6njkmRMxE`
	pref := tele.Settings{
		Token:  (token),
		Poller: &tele.LongPoller{Timeout: 5 * time.Second},
	}
	log.Println("Setting up the bot...")
	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err) // shutting down, the server is useless without the bot
		return
	}

	bot.Handle("/schedule", func(c tele.Context) error {
		return c.Send("Данная функция ещё не написана!")
	})

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Checking if the schedule has changed...")
		resp, err := getSchedule()
		if err == nil {
			defer resp.Body.Close()
		} else {
			log.Println(err)
			continue
		}
		saveFile(filepathnew, resp)
		if !equalFiles(filepathnew, filepathold) {
			log.Println("The schedule has changed!")
			err := os.Remove(filepathold)
			checkerr(err, "Error removing the new schedule file: ")
			err = os.Rename(filepathnew, filepathold)
			checkerr(err, "Error renaming downloaded schedule file: ")
			sendable, err := os.Open(filepathold)
			checkerr(err, "Error opening downloaded schedule file: ")
			if err == nil {
				// Create the document
				doc := &tele.Document{
					File:     tele.FromReader(sendable),
					FileName: "schedule.pdf",
				}
				// Send the document
				_, err = bot.Send(&tele.Chat{ID: -1001778030528}, doc)
				checkerr(err)
			} else {
				log.Println(err)
				continue
			}
		} else {
			log.Println("The schedule stays the same.")
			// got the schedule, but it is the same
			err := os.Remove(filepathnew)
			checkerr(err, "Error removing the new schedule file: ")
		}
	}
	bot.Start()

}

func equalFiles(file1, file2 string) bool {
	// per comment, better to not read an entire file into memory
	// this is simply a trivial example.
	f1, err := ioutil.ReadFile(file1)

	if err != nil {
		log.Fatal(err)
	}

	f2, err := ioutil.ReadFile(file2)

	if err != nil {
		log.Fatal(err)
	}

	return (bytes.Equal(f1, f2)) // Per comment, this is significantly more performant.
}

func getSchedule() (*http.Response, error) {
	resp, err := http.Get("https://biit39.ru/raspisanie-zanyatiy/")
	if err != nil {
		return nil, fmt.Errorf("Error while fetching the schedule link:", err)
		//fmt.Println("Error while fetching the schedule link:", err)
		//continue
	}
	defer resp.Body.Close()

	// Search for the schedule link in the HTML
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`https\:\/\/biit39\.ru\/wp-content/uploads/\d{4}\/\d{2}\/raspisanie-ochno-zaochnoe.{0,4}\.pdf`)
	// Yes, regex are slow and inefficient, so what? Sue me.
	// Finding all matches in the body of the response
	matches := re.FindAllString(string(body), -1)
	link := ""
	if len(matches) > 0 {
		link = matches[0]
	} else {
		//log.Println("Cannot retrieve the link")
		//contnue
		return nil, fmt.Errorf("Cannot retrieve the link")
	}
	// Download the schedule file
	resp, err = http.Get(link)
	if err != nil {
		//log.Println("Error while downloading the schedule file: ", err)
		return nil, fmt.Errorf("Error while downloading the schedule file: ", err)
		//continue
	}
	// defer resp.Body.Close()
	return resp, nil
}

func saveFile(filePath string, resp *http.Response) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
		//	continue
	}
	_, err = io.Copy(file, resp.Body)
	//defer
	file.Close()
	if err != nil {
		return err
		//continue
	} else {
		return nil
	}
}

func planner() {
	ticker := time.NewTicker(8 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
	}
}

func checkerr(err error, msg ...string) {
	//fmt.Printf("checkerr %v ; %v\n", err, msg)
	if err != nil {
		if len(msg) > 0 {
			log.Printf("%v: %v\n", err, msg[0])
		} else {
			log.Print(err)
		}
	}
}
