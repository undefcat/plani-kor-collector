package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var sc *bufio.Scanner

func init() {
	sc = bufio.NewScanner(os.Stdin)
}

func scanString() string {
	sc.Scan()
	return sc.Text()
}

func main() {
	fmt.Println("__() 수집 프로그램")
	fmt.Println("-----------------")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var rdPath string

	for {
		fmt.Printf("읽을 경로를 입력하세요. 현재경로: %s\n", wd)
		fmt.Printf("경로: ")

		// 루트 디렉터리
		rdPath, err = filepath.Abs(scanString())
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// 디렉터리 존재유무 확인
		f, err := os.Stat(rdPath)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				fmt.Printf("%s 경로는 존재하지 않습니다.\n", rdPath)

			default:
				fmt.Printf("%s", err.Error())
			}

			continue
		}

		// 디렉터리 여부 확인
		if ! f.IsDir() {
			fmt.Printf("%s 경로는 디렉터리가 아닙니다.\n", rdPath)
		}

		break
	}

	var out *os.File
	for {
		fmt.Printf("출력 파일 경로를 입력하세요(파일 이름 포함). 현재경로: %s\n", wd)
		fmt.Printf("경로: ")

		// 출력 디렉터리
		outPath, err := filepath.Abs(scanString())
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		_, err = os.Stat(outPath)
		if err == nil {
			fmt.Printf("%s 는 이미 존재하는 파일입니다. 다른 경로를 선택해주세요.\n", outPath)
			continue
		}

		// 디렉터리 없는 경우 생성하고
		// 다음 단계로 넘어간다.
		if os.IsNotExist(err) {
			out, err = os.Create(outPath)
			if err == nil {
				defer out.Close()

				break
			}

			fmt.Println(err.Error())
			continue
		}

		break
	}

	target := regexp.MustCompile(`__\('(.+?)'\)`)

	fmt.Println("작업을 시작합니다.")

	countChan := make(chan int, 3)

	var wait sync.WaitGroup

	wait.Add(1)
	go func() {
		// 루트 디렉터리부터 순회
		err = filepath.Walk(rdPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				// continue
				return err
			}

			// 숨김 파일은 지나간다.
			name := filepath.Base(path)
			if name[0] == '.' {
				return err
			}

			// 파일을 읽는다.
			strs, err1 := ioutil.ReadFile(path)
			if err1 != nil {
				log.Println(err1)
				return err1
			}

			str := string(strs)

			// __('')를 모두 찾는다.
			matches := target.FindAllStringSubmatch(str, -1)

			// 경로,한글 형식으로 바꾼다.
			results := make([]string, len(matches))

			for mi := range matches {
				result := fmt.Sprintf("\"%s\",\"%s\"\n", path, matches[mi][1])
				results[mi] = result
			}

			out.WriteString(strings.Join(results, ""))

			// 완료된 파일작업 진행 현황을 알린다.
			countChan <- len(results)
			return err
		})

		if err != nil {
			fmt.Println(err.Error())
		}

		wait.Done()
		close(countChan)
	}()

	wait.Add(1)
	go func() {
		total := 0
		for v := range countChan {
			total += v
			fmt.Printf("\r현재 %v개의 단어를 추출했습니다.", total)
		}

		fmt.Printf("\r총 %v개의 단어를 추출했습니다.", total)
		wait.Done()
	}()

	wait.Wait()
}
