package prt

import (
    "porto" // local link to "github.com/yandex/porto/src/api/go/porto"
    "strconv"
)

func MakePortoApi() (porto.API, error) {
    return porto.Connect()
}

func MakePortoContainer(api porto.API, name string) error {
    if err := api.CreateWeak(name); err != nil {
        return err
    }

    defer func() {
        if r := recover(); r != nil {
            api.Close()
        }
    }()

    if err := api.SetProperty(name, "command", "bash -c 'while [[ 1 ]]; do sleep 3; done'"); err != nil {
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

func GetPortoCpu(api porto.API, name string) (uint64, error) {
    val, err := GetPortoProperty(api, name, "cpu_usage")
    if err != nil {
        return 0, err
    }

    return strconv.ParseUint(val, 10, 64)
}

func CleanupPortoContainer(api porto.API, name string) (err error) {
    if err = api.Stop(name); err != nil {
        return
    }
    return
}
