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

func GetPortoCpu(api porto.API, name string) uint64 {
    val, err := GetPortoProperty(api, name, "cpu_usage")
    if err != nil {
        panic(err)
    }

    var intVal uint64
    intVal, err = strconv.ParseUint(val, 10, 64)
    if err != nil {
        panic(err)
    }

    return intVal
}

func CleanupPortoContainer(api porto.API, name string) (err error) {
    if err = api.Stop(name); err != nil {
        return
    }
    return
}
