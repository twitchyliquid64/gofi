gofi
============

*A lightweight, raspberry-pi friendly controller for Ubiquity Access Points*

Go-fi is easy to build and deploy, providing basic network management
features with very tiny CPU/Memory/Disk requirements.


Is this legal?
--------------

I emailed Ubiquiti and they said they were fine with gofi being publically available, and that they are aware of many similar projects. So go for it!


How do I build?
----------------

```shell
git clone https://github.com/twitchyliquid64/gofi
cd gofi
export GOPATH=`pwd`

# A very basic, stateless controller
go build -o statelessController gofi/controllers/stateless
```

How do I run?
---------------

There are two controllers available, statelessController and basicController.

Stateless controller is super simple, you start it on the command line and it will adopt all APs. You pass in network information on the command line, and it
will configure ALL APs to use those. It is stateless, so it will leave the APs with default credentials, and if it is ever restarted re-adopt and reconfigure them.

Usage:

```
Usage of ./statelessController:
  -addr string
    	Controller LAN IP - autodetected if not set
  -enable_5g
    	Make network available on 5G as well as 2.4G (default true)
  -enable_bandsteering
    	Steer clients to 5G network
  -pw string
    	Network password (default "fiog")
  -ssid string
    	Network name (default "gofi")
```

Example:

```./statelessController -enable_5g -enable_bandsteering -ssid "silly_example" -pw "mynetworkpassword"```


LICENSE (MIT)
--------------

Copyright (c) 2017 twitchyliquid64

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
