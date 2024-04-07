package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/achetronic/tapogo/pkg/tapogo"
	"github.com/choria-io/fisk"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	ip           net.IP
	user         string
	pass         string
	jsonFormat   bool
	choriaFormat bool
	labels       map[string]string
)

func main() {
	labels = make(map[string]string)

	tapoctl := fisk.New("tapoctl", "Controls TP-Link Tapo Smart Plugs")

	addopts := func(c *fisk.CmdClause) {
		c.Arg("address", "Device IP address").Envar("TAPO_ADDRESS").Required().IPVar(&ip)
		c.Arg("username", "Device username").Envar("TAPO_USER").Required().StringVar(&user)
		c.Arg("password", "Device password").Envar("TAPO_PASSWORD").Required().StringVar(&pass)
	}

	addopts(tapoctl.Command("on", "Turns the device on").Action(onCommand))
	addopts(tapoctl.Command("off", "Turns the device off").Action(offCommand))
	info := tapoctl.Command("info", "Shows device information").Action(infoCommand)
	addopts(info)
	info.Flag("json", "Produce JSON output").UnNegatableBoolVar(&jsonFormat)

	energy := tapoctl.Command("energy", "Retreives device energy usage statistics").Action(energyCommand)
	addopts(energy)
	energy.Flag("json", "Produce JSON output").UnNegatableBoolVar(&jsonFormat)
	energy.Flag("choria", "Produce Choria Metric output").UnNegatableBoolVar(&choriaFormat)
	energy.Flag("label", "Labels to apply to Choria Metric output").StringMapVar(&labels)

	tapoctl.MustParseWithUsage(os.Args[1:])
}

func plug() (*tapogo.Tapo, error) {
	return tapogo.NewTapo(ip.String(), user, pass, &tapogo.TapoOptions{})
}

func onCommand(_ *fisk.ParseContext) error {
	p, err := plug()
	if err != nil {
		return err
	}

	_, err = p.TurnOn()
	if err != nil {
		return err
	}

	nfo, err := p.DeviceInfo()
	if err != nil {
		return err
	}

	if !nfo.Result.DeviceOn {
		return fmt.Errorf("device failed to power on for an unknown reason")
	}

	fmt.Println("Powered on")
	return nil
}

func offCommand(_ *fisk.ParseContext) error {
	p, err := plug()
	if err != nil {
		return err
	}

	_, err = p.TurnOff()
	if err != nil {
		return err
	}

	nfo, err := p.DeviceInfo()
	if err != nil {
		return err
	}

	if nfo.Result.DeviceOn {
		return fmt.Errorf("device failed to power off for an unknown reason")
	}

	fmt.Println("Powered off")
	return nil
}

func infoCommand(_ *fisk.ParseContext) error {
	p, err := plug()
	if err != nil {
		return err
	}

	nfo, err := p.DeviceInfo()
	if err != nil {
		return err
	}

	if jsonFormat {
		j, err := json.MarshalIndent(nfo.Result, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(j))
		return nil
	}

	state := "Off"
	if nfo.Result.DeviceOn {
		state = "On"
	}

	fmt.Println("Device Information:")
	fmt.Println()
	name, err := base64.StdEncoding.DecodeString(nfo.Result.Nickname)
	if err == nil {
		fmt.Printf("         Nick Name: %s\n", name)
	}
	fmt.Printf("              Icon: %s\n", nfo.Result.Avatar)
	fmt.Printf("       Power State: %s\n", state)
	fmt.Printf("              Type: %s %s\n", nfo.Result.Type, nfo.Result.Model)
	fmt.Printf("         Device ID: %s\n", nfo.Result.DeviceId)
	fmt.Printf("  Firmware Version: %s\n", nfo.Result.FwVer)
	fmt.Printf("  Hardware Version: %s\n", nfo.Result.HwVer)
	fmt.Printf("            Region: %s\n", nfo.Result.Region)
	fmt.Println()
	fmt.Println("Network Information:")
	fmt.Println()
	fmt.Printf("        IP Address: %s\n", nfo.Result.Ip)
	fmt.Printf("       MAC Address: %s\n", nfo.Result.Mac)
	ssid, err := base64.StdEncoding.DecodeString(nfo.Result.Ssid)
	if err == nil {
		fmt.Printf("         WiFi SSID: %s\n", ssid)
	}
	fmt.Printf("        RSSI Level: %d\n", nfo.Result.Rssi)
	fmt.Printf("      Signal Level: %d\n", nfo.Result.SignalLevel)

	return nil
}

func energyCommand(_ *fisk.ParseContext) error {
	p, err := plug()
	if err != nil {
		return err
	}

	nfo, err := p.GetEnergyUsage()
	if err != nil {
		return err
	}

	currentPower := float64(nfo.Result.CurrentPower) / 1000
	todayEnergy := float64(nfo.Result.TodayEnergy) / 1000
	monthEnergy := float64(nfo.Result.MonthEnergy) / 1000

	switch {
	case jsonFormat:
		j, err := json.MarshalIndent(nfo.Result, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(j))
		return nil

	case choriaFormat:
		data := map[string]any{
			"labels": labels,
			"metrics": map[string]any{
				"current_power_watt":    currentPower,
				"today_energy_kwh":      todayEnergy,
				"month_energy_kwh":      monthEnergy,
				"today_runtime_seconds": nfo.Result.TodayRuntime,
				"month_runtime_seconds": nfo.Result.MonthRuntime,
			}}
		j, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(j))
	default:
		fmt.Println("Power Usage")
		fmt.Println()
		fmt.Printf("    Current Power: %.3fW\n", currentPower)
		fmt.Printf("     Today Energy: %.3fkWh\n", todayEnergy)
		fmt.Printf("     Month Energy: %.3fkWh\n", monthEnergy)
		fmt.Printf("    Today Runtime: %s\n", secondsToHuman(nfo.Result.TodayRuntime))
		fmt.Printf("    Month Runtime: %s\n", secondsToHuman(nfo.Result.TodayRuntime))
	}

	return nil

}

func plural(count int, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}
	return
}

func secondsToHuman(input int) (result string) {
	years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
	seconds = input % (60 * 60 * 24 * 7 * 30)
	weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
	seconds = input % (60 * 60 * 24 * 7)
	days := math.Floor(float64(seconds) / 60 / 60 / 24)
	seconds = input % (60 * 60 * 24)
	hours := math.Floor(float64(seconds) / 60 / 60)
	seconds = input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = input % 60

	if years > 0 {
		result = plural(int(years), "year") + plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if months > 0 {
		result = plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if weeks > 0 {
		result = plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if days > 0 {
		result = plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if hours > 0 {
		result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if minutes > 0 {
		result = plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else {
		result = plural(int(seconds), "second")
	}

	return strings.Trim(result, " ")
}
