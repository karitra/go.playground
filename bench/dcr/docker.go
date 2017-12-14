//
// Notes:
//
//  docker/docker/client is using socket desc per connection
//
//
package dcr

import (
    docker "github.com/docker/docker/client"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"

    "golang.org/x/net/context"
    "io"

    "encoding/json"
)

const (
    dockerImage = "ubuntu"
)

func MakeDockerClient() (cli *docker.Client, err error) {
    cli, err = docker.NewEnvClient()
    return
}


func MakeDockerContainer(ctx context.Context, cli *docker.Client, name string) (id string, err error) {
    conf := container.Config{}

    conf.Image = dockerImage
    conf.Tty = true

    hostConf := container.HostConfig{}
    hostConf.AutoRemove = true

    cont, err := cli.ContainerCreate(ctx, &conf, &hostConf, nil, name)
    if err != nil {
        return
    }

    id = cont.ID
    options := types.ContainerStartOptions{}
    cli.ContainerStart(ctx, id, options)

    return
}


func CleanupDockerContainer(ctx context.Context, cli *docker.Client, id string) {
    if err := cli.ContainerKill(ctx, id, "KILL"); err != nil {
        panic(err)
    }
    waitCh, errCh := cli.ContainerWait(ctx, id, container.WaitConditionRemoved)
    select {
        case <- waitCh:
            break
        case err := <- errCh:
            panic(err)
    }
}


func GetDockerStatsStream(ctx context.Context, client *docker.Client, id string) (out io.ReadCloser) {
    stats, err := client.ContainerStats(ctx, id, true)
    if err != nil {
        panic(err)
    }
    return stats.Body
}


func ParseDockerStats(dec *json.Decoder) (out types.Stats) {
    if dec.More() {
        if err := dec.Decode(&out); err != nil {
            panic(err)
        }
    } else {
        panic("fuh")
    }

    return
}


func GetDockerStats(ctx context.Context, client *docker.Client, id string) (out types.Stats) {
    stats, err := client.ContainerStats(ctx, id, false)
    if err != nil {
        panic(err)
    }

    defer stats.Body.Close()

    dec := json.NewDecoder(stats.Body)
    if err := dec.Decode(&out); err != nil {
        panic(err)
    }

    return
}
