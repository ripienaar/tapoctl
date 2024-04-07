## Control for Tapo P1xx Plugs

A small client utility for interacting with the Tapo plugs running the KLAP protocol.

Tested on a Raspberry PI with Go 1.22

## Usage?

Basic usage information:

```nohighlight
$ tapoctl --help
usage: tapoctl [<flags>] <command> [<args> ...]

Controls TP-Link Tapo Smart Plugs

Commands:
  on      Turns the device on
  off     Turns the device off
  info    Shows device information
  energy  Retreives device energy usage statistics

Global Flags:
  --help  Show context-sensitive help
```

All commands take `<address> <username> <password>` arguments, these can be set in the environment with `TAPO_ADDRESS`, `TAPO_USER` and `TAPO_PASSWORD`.

Obtain device info:

```nohighlight
$ tapoctl info 192.168.1.10
Device Information:

         Nick Name: Office
              Icon: plug
       Power State: On
              Type: SMART.TAPOPLUG P110
         Device ID: XXX
  Firmware Version: 1.3.0 Build 230905 Rel.152200
  Hardware Version: 1.0
            Region: Europe/Malta

Network Information:

        IP Address: 192.1.10
       MAC Address: 9C-9C-9C-9C-9C-9C
         WiFi SSID: example
        RSSI Level: -47
      Signal Level: 3
```

Read energy values:

```nohighlight
$ tapoctl energy 192.168.1.10
Power Usage

    Current Power: 20.831W
     Today Energy: 0.010kWh
     Month Energy: 0.010kWh
    Today Runtime: 1 minute 45 seconds
    Month Runtime: 1 minute 45 seconds
```

It supports JSON output:

```nohighlight
$ tapoctl energy 192.168.1.10 --json
{
  "today_runtime": 106,
  "month_runtime": 106,
  "today_energy": 10,
  "month_energy": 10,
  "local_time": "2024-04-07 13:37:00",
  "electricity_charge": [
    0,
    0,
    0
  ],
  "current_power": 20701
}
```

And also the format required by Choria Metric watchers:

```
$ tapoctl energy --choria --label location:office
{
  "labels": {
    "location": "office"
  },
  "metrics": {
    "current_power_watt": 20.777,
    "month_energy_kwh": 0.011,
    "month_runtime_seconds": 107,
    "today_energy_kwh": 0.011,
    "today_runtime_seconds": 107
  }
}
```

## Contact?

R.I. Pienaar / rip@devco.net / [devco.net](https://www.devco.net/)