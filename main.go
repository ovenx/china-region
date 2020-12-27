package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)
type Province struct  {
	Name string `json:"name"`
	Code string `json:"code"`
	CityList map[string]City `json:"city_list"`
}

type City struct {
	Name string `json:"name"`
	Code string `json:"code"`
	AreaList map[string]Area `json:"area_list"`
}

type Area struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func main () {
	link := "http://www.mca.gov.cn/article/sj/xzqh/2020/2020/2020112010001.html"
	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	data := make(map[string]Province, 0)
	province := Province{}
	city := City{}
	area := Area {}
	provinceCode := ""
	doc.Find("tr[height=\"19\"]").Each(func(i int, s *goquery.Selection) {
		code := ""
		name := ""
		s.Find("td").Each(func(j int, s *goquery.Selection) {
			if j == 1 {
				code = s.Text()
			} else if j == 2 {
				strEncode := url.QueryEscape(s.Text())
				strEncode = strings.Replace(strEncode, "%C2%A0", "%20", -1)
				// strEncode = strings.Replace(strEncode, "+", "", -1)
				name, _ = url.QueryUnescape(strEncode)
			} else {
				return
			}
		})
		if name[: 1] != " " {
			province = Province{Code: code, Name: strings.Replace(name, " ", "", -1), CityList: map[string]City{}}
			provinceCode = code
			city = City{}
			area = Area{}
		} else if name[1:2] != " " {
			city = City{Code: code, Name: strings.Replace(name, " ", "", -1), AreaList: map[string]Area{}}
			province.CityList[code] = city
		} else {
			if reflect.DeepEqual(city, City{}){
				cityCode, _ := strconv.Atoi(provinceCode)
				cityCodeStr := strconv.Itoa(cityCode + 100)
				city = City{Code: cityCodeStr, Name: province.Name, AreaList: map[string]Area{}}
				province.CityList[cityCodeStr] = city
			}
			area = Area{Code: code, Name: strings.Replace(name, " ", "", -1)}
			city.AreaList[code] = area
		}

		data[provinceCode] = province
	})
	// fmt.Println(data)
	jsonPath := "region.json"
	jsonData,_ := json.MarshalIndent(&data, "", "  ")
	fmt.Println(string(jsonData))
	WriteToFile(jsonPath, string(jsonData))
}

func WriteToFile(filePath string, str string) {
	out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open file fail:", err)
		return
	}
	defer out.Close()
	writer := bufio.NewWriter(out)
	writer.WriteString(str)
	writer.Flush()
}
