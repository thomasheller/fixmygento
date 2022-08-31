package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Strategy struct {
	Cmds []string
}

func (s Strategy) Name() string {
	return strings.Join(s.Cmds, ", ")
}

func (s Strategy) Run() error {
	for _, cmd := range s.Cmds {
		if err := s.runExternalCmd(cmd); err != nil {
			return err
		}
	}

	return nil
}

func (s Strategy) runExternalCmd(cmd string) error {
	const dc = "docker-compose"
	args := []string{"exec", "-T", "fpm", "bin/magento", cmd}

	log.Printf("Running: %s %s", dc, strings.Join(args, " "))

	c := exec.Command(dc, args...)

	stdoutStderr, err := c.CombinedOutput()

	fmt.Printf("%s", stdoutStderr)

	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("Attempting to fix your 'gento... 🧑‍🔧")

	playAudio()

	strategies := []Strategy{
		// run cache:flush first...
		Strategy{[]string{
			"cache:flush",
			"setup:di:compile",
			"setup:upgrade",
			"indexer:reindex",
		}},
		Strategy{[]string{
			"cache:flush",
			"setup:upgrade",
			"setup:di:compile",
			"indexer:reindex",
		}},
		// run cache:flush later... just to make sure
		Strategy{[]string{
			"setup:di:compile",
			"setup:upgrade",
			"cache:flush",
			"indexer:reindex",
		}},
		Strategy{[]string{
			"setup:upgrade",
			"setup:di:compile",
			"cache:flush",
			"indexer:reindex",
		}},
		// run reindexing first...?!
		Strategy{[]string{
			"indexer:reindex",
			"cache:flush",
			"setup:di:compile",
			"setup:upgrade",
		}},
		Strategy{[]string{
			"indexer:reindex",
			"cache:flush",
			"setup:upgrade",
			"setup:di:compile",
		}},
		Strategy{[]string{
			"indexer:reindex",
			"setup:di:compile",
			"setup:upgrade",
			"cache:flush",
		}},
		Strategy{[]string{
			"indexer:reindex",
			"setup:upgrade",
			"setup:di:compile",
			"cache:flush",
		}},
		// run cache:flush between setup commands...
		Strategy{[]string{
			"setup:di:compile",
			"cache:flush",
			"setup:upgrade",
			"indexer:reindex",
		}},
		Strategy{[]string{
			"setup:upgrade",
			"cache:flush",
			"setup:di:compile",
			"indexer:reindex",
		}},
		Strategy{[]string{
			"indexer:reindex",
			"setup:di:compile",
			"cache:flush",
			"setup:upgrade",
		}},
		Strategy{[]string{
			"indexer:reindex",
			"setup:upgrade",
			"cache:flush",
			"setup:di:compile",
		}},
		// flush cache before reindexing...?
		Strategy{[]string{
			"cache:flush",
			"indexer:reindex",
			"setup:di:compile",
			"setup:upgrade",
		}},
		Strategy{[]string{
			"cache:flush",
			"indexer:reindex",
			"setup:upgrade",
			"setup:di:compile",
		}},
		// run reindexing between setup commands...?!!
		Strategy{[]string{
			"cache:flush",
			"setup:di:compile",
			"indexer:reindex",
			"setup:upgrade",
		}},
		Strategy{[]string{
			"cache:flush",
			"setup:upgrade",
			"indexer:reindex",
			"setup:di:compile",
		}},
		Strategy{[]string{
			"setup:di:compile",
			"indexer:reindex",
			"setup:upgrade",
			"cache:flush",
		}},
		Strategy{[]string{
			"setup:upgrade",
			"indexer:reindex",
			"setup:di:compile",
			"cache:flush",
		}},
		// setup, reindex, flush...
		Strategy{[]string{
			"setup:di:compile",
			"setup:upgrade",
			"indexer:reindex",
			"cache:flush",
		}},
		Strategy{[]string{
			"setup:upgrade",
			"setup:di:compile",
			"indexer:reindex",
			"cache:flush",
		}},
		// remaining obscure permutations...
		Strategy{[]string{
			"setup:di:compile",
			"cache:flush",
			"indexer:reindex",
			"setup:upgrade",
		}},
		Strategy{[]string{
			"setup:di:compile",
			"indexer:reindex",
			"cache:flush",
			"setup:upgrade",
		}},
		Strategy{[]string{
			"setup:upgrade",
			"cache:flush",
			"indexer:reindex",
			"setup:di:compile",
		}},
		Strategy{[]string{
			"setup:upgrade",
			"indexer:reindex",
			"cache:flush",
			"setup:di:compile",
		}},
	}

	for i, s := range strategies {
		log.Printf(`Attempting strategy %d/%d: "%s"...`, i+1, len(strategies), s.Name())

		err := s.Run()

		if err2 := logMessage(s.Name(), err == nil); err2 != nil {
			log.Printf("Warning: Failed to log result: %s", err2)
		}

		if err != nil {
			log.Printf("Error: %s", err)
			log.Println("Failure ❌")
			continue
		}

		log.Println("Success! ✅")

		return
	}

	log.Fatalf("I'm sorry, none of the strategies worked. 🙁")
}

func playAudio() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("🙁 Failed to play audio: Failed to find home directory: %s", err)

		return
	}

	audioFile := filepath.Join(homeDir, ".fixmygento", "loop.mp3")

	f, err := os.Open(audioFile)
	if err != nil {
		log.Printf("🙁 Failed to play audio: Could not open audio file \"%s\": %s", err)

		return
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Printf("🙁 Failed to play audio: Error decoding MP3 file: %s", err)

		return
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	loop := beep.Loop(-1, streamer)
	done := make(chan bool)
	speaker.Play(beep.Seq(loop, beep.Callback(func() {
		done <- true
	})))
}

func logMessage(name string, success bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current working directory: %s", err)
	}

	var result string

	if success {
		result = "success"
	} else {
		result = "failure"
	}

	msg := fmt.Sprintf("[%s] %s! %s @ %s\n", time.Now().Format(time.RFC3339), result, name, wd)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Failed to get user's home directory: %s", err)
	}

	file := filepath.Join(home, ".fixmygento.log")

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(`Failed to open logfile "%s" for writing: %s`, file, err)
	}

	defer f.Close()

	if _, err := f.WriteString(msg); err != nil {
		return fmt.Errorf(`Failed to append to logfile "%s": %s`, file, err)
	}

	return nil
}
