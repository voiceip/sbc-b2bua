package sbc

import (
    "sippy/conf"
    "sippy/net"
)


type Config struct {
    sippy_conf.Config

    NH_addr *sippy_net.HostPort
}