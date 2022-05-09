package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func failure(w http.ResponseWriter, message string) {
	log.Println(message)
	http.Error(w, message, http.StatusInternalServerError)
}

func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", *bckPtr, nil)
	if err != nil {
		failure(w, "Failed to compose request")
		return
	}
	if len(*usrPtr) > 0 && len(*pwdPtr) > 0 {
		req.SetBasicAuth(*usrPtr, *pwdPtr)
	}
	resp, err := client.Do(req)
	if err != nil {
		failure(w, "Failed to send request to upstream")
		return
	}

	if resp.StatusCode != 200 {
		failure(w, "Failed talking to upstream")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		failure(w, "Failed to receive responce from upstream")
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	sb := string(body)
	list := strings.Split(strings.ReplaceAll(sb, "\r\n", "\n"), "\n")
	helpPrinted := false

	for _, v := range list {
		if len(v) > 0 {
			if strings.Count(v, "=") == 1 {
				s := strings.Split(v, "=")
				n, err := strconv.ParseInt(strings.TrimSpace(s[1]), 10, 64)
				if err == nil {
					a := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s[0], ".", "_"), ":", "_"), "%", "p")))
					s := strings.Split(a, "_")
					l := len(s)
					if l >= 3 {
						l -= 1

						//fmt.Println("Input=>", a, "=>", s[l-2], "=>", s[l-1], "=>", s[l], strings.Compare(s[l-1], "time"), strings.Compare(s[l], "avg"), n)

						cName := ""
						vName := ""

						if (strings.Compare(s[l-1], "time") == 0) && (strings.Compare(s[l], "avg") == 0) {
							//fmt.Println(1)
							cName = strings.Join(s[:l-1], "_")
							vName = "time_avg"

						} else if strings.Compare(s[l-2], "time") == 0 && strings.Compare(s[l-1], "avg") == 0 && strings.Compare(s[l], "mov") == 0 {
							//fmt.Println(2)
							cName = strings.Join(s[:l-3], "_")
							vName = "time_avg_mov"

						} else if strings.Compare(s[l], "agg") == 0 {
							//fmt.Println(3)
							cName = strings.Join(s[:l-2], "_")
							vName = s[l-1] + "_agg"

						} else if strings.Compare(s[l-1], "http") == 0 {
							//fmt.Println(4)
							cName = strings.Join(s[:l-1], "_")
							vName = "http_" + s[l]
						} else {
							//fmt.Println(5)
							cName = strings.Join(s[:l], "_")
							vName = s[l]
						}

						c := strings.SplitN(cName, "_", 2)
						if len(c) == 2 {

							cName = c[0]
							bName := c[1]

							// fmt.Println("Output=>", a, s, vName, "cName=>", cName, "oName=>", oName)
							if helpPrinted != false {
								helpPrinted = true
								fmt.Fprint(w, "# HELP ", *pfxPtr, " Gauge value", "\n")
								fmt.Fprint(w, "# TYPE ", *pfxPtr, " gauge\n")
							}
							fmt.Fprint(w, *pfxPtr, "{pool=\"", cName, "\",back=\"", bName, "\",gauge=\"", vName, "\"} ", n, "\n")
						}
					}
				} else {
					fmt.Println(err)
				}
			}
		}
	}
}
