package main

import (
    "encoding/xml"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"
    "fmt"
    "strings"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "golang.org/x/text/encoding/charmap"
    "gopkg.in/yaml.v2"
)


type Config struct {
    ValueMapping map[string]string `yaml:"value_mapping"`
}

type ValCurs struct {
    Valutes []Valute `xml:"Valute"`
}

type Valute struct {
    CharCode string `xml:"CharCode"`
    Nominal  string `xml:"Nominal"`
    Value    string `xml:"Value"`
}

var (
    exchangeRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "exchange_rate",
        Help: "Exchange rate",
    }, []string{"code", "name"})
    config Config
)

func main() {
    addr := getEnv("LISTEN_ADDR", "0.0.0.0:9393")
    configPath := getEnv("CONFIG_FILE", "./config.yaml")
    interval, _ := strconv.Atoi(getEnv("REFRESH_INTERVAL", "600"))

    log.Println("Listening on http://", addr)

    if configPath != "" {
        file, err := ioutil.ReadFile(configPath)
        if err == nil {
            yaml.Unmarshal(file, &config)
        }
    }

    prometheus.MustRegister(exchangeRate)

    go func() {
        for {
            log.Println("Fetching exchange rates...")
            getExchangeRate()
            log.Println("Sleeping for", interval, "seconds")
            time.Sleep(time.Duration(interval) * time.Second)
        }
    }()

    r := mux.NewRouter()
    r.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(addr, r))
}

func getEnv(key string, defaultVal string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return defaultVal
}

func getExchangeRate() {
    client := &http.Client{}

    // Get current date and format it
    now := time.Now()
    dateReq := now.Format("02/01/2006")

    req, _ := http.NewRequest("GET", "http://www.cbr.ru/scripts/XML_daily.asp?date_req="+dateReq, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
    req.Header.Set("Referer", "http://www.cbr.ru/development/SXML/")
    resp, err := client.Do(req)
    if err != nil {
        log.Println("Error fetching data:", err)
        return
    }
    defer resp.Body.Close()

    decoder := xml.NewDecoder(resp.Body)
    decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
        switch charset {
        case "windows-1251":
            return charmap.Windows1251.NewDecoder().Reader(input), nil
        default:
            return nil, fmt.Errorf("unknown charset: %s", charset)
        }
    }

    var valCurs ValCurs
    err = decoder.Decode(&valCurs)
    if err != nil {
        log.Println("Error unmarshalling XML:", err)
        return
    }

    for _, v := range valCurs.Valutes {
        // Replace commas with periods
        nominalStr := strings.ReplaceAll(v.Nominal, ",", ".")
        valueStr := strings.ReplaceAll(v.Value, ",", ".")

        nominal, _ := strconv.ParseFloat(nominalStr, 64)
        value, _ := strconv.ParseFloat(valueStr, 64)
        rate := value / nominal

        name, ok := config.ValueMapping[v.CharCode]
        if !ok {
            name = ""
        }

        log.Println("Setting exchange_rate{code=\"" + v.CharCode + "\", name=\"" + name + "\"} = ", rate)
        exchangeRate.WithLabelValues(v.CharCode, name).Set(rate)
    }
}
