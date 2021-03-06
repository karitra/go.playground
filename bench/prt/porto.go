package prt

import (
    "porto" // local link to "github.com/yandex/porto/src/api/go/porto"
    "strconv"
    "strings"
    "fmt"
    "net"
    "math/rand"
    "time"
    "regexp"
)

const toTry = 10
const randSleepIvalMs = 100

var (
    spacesRegexp, _ = regexp.Compile("[ ]+")
)

func MakePortoApiWithPanic() (api porto.API) {
    apiCh, errCh := MakePortoApi()

    select {
    case api = <- apiCh:
    case err := <- errCh: panic(err)
    }
    return
}

type NetIfStat struct {
    Name string
    BytesCount uint64
}

const PAIR_LEN = 2
const (
    pairName = iota
    pairVal
)

func parseNameUIntValPair(eth string) (nstat NetIfStat, err error) {
    pair := strings.Split(eth, ": ")
    if len(pair) == PAIR_LEN {

        var v uint64
        v, err = strconv.ParseUint(pair[pairVal], 10, 64)
        if err != nil {
            return
        }

        name := strings.Trim(pair[pairName], " ")
        name  = spacesRegexp.ReplaceAllString(name, "_")

        nstat = NetIfStat{
            Name: name,
            BytesCount: v }

    } else {
        err = fmt.Errorf("failed to parse net record")
    }

    return
}


func ParseNetValues(val string) (ifs []NetIfStat) {
    for _, eth := range strings.Split(val, ";") {
        nf, err := parseNameUIntValPair(eth)
        if err == nil {
            ifs = append(ifs, nf)
        }
    }

    return
}


func MakePortoApi() (apiCh chan porto.API, errCh chan error) {
    apiCh = make(chan porto.API)
    errCh = make(chan error)

    go func() {
        for attempts := 0; attempts < toTry; attempts++ {
            api, err := porto.Connect()
            if err != nil {
                if err, ok := err.(net.Error); ok && err.Temporary() {
                    toSleep := time.Duration(rand.Uint64() % randSleepIvalMs * uint64(time.Millisecond))
                    time.Sleep(toSleep)
                    continue
                } else {
                    errCh <- err
                }
            }

            apiCh <- api
            return
        }
    }()

    return apiCh, errCh
}


func MakePortoContainer(api porto.API, name string) error {
    if err := api.CreateWeak(name); err != nil {
        fmt.Printf("error1 %v for name %s", err, name)
        return err
    }

    cmd := "bash -c 'while [[ 1 ]]; do sleep 2 && ping -c 3 ya.ru; done'"
    if err := api.SetProperty(name, "command", cmd); err != nil {
       fmt.Printf("error2 %v", err)
       panic(err)
    }

    return nil
}

func RunPortoContainer(api porto.API, name string) error {
    return api.Start(name)
}

func GetPortoProperty(api porto.API, name string, prop string) (string, error) {
    return api.GetProperty(name, prop)
}

func GetUint64Property(api porto.API, name string, propName string) (intVal uint64, err error) {
    val, err := GetPortoProperty(api, name, propName)
    if err != nil {
        return 0, err
    }

    intVal, err = strconv.ParseUint(val, 10, 64)
    if err != nil {
        return intVal, err
    }

    return intVal, err
}

func GetPortoCpu(api porto.API, name string) (intVal uint64, err error) {
    return GetUint64Property(api, name, "cpu_usage")
}

func GetPortoMem(api porto.API, name string) (intVal uint64, err error) {
    return GetUint64Property(api, name, "memory_usage")
}

func CleanupPortoContainer(api porto.API, name string) (err error) {
    if err = api.Stop(name); err != nil {
        return
    }
    return
}
