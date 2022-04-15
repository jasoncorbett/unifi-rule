package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/namsral/flag"
)

func main() {
	var err error
	unifiIp := "10.0.0.1"
	unifiUsername := ""
	unifiPassword := ""
	unifiRuleNumber := ""

	flag.StringVar(&unifiIp, "ip", "192.168.1.1", "The ip address of the unifi dream machine")
	flag.StringVar(&unifiUsername, "unifi-username", "username", "The username to log into the controller with (you can use environment variables)")
	flag.StringVar(&unifiPassword, "unifi-password", "password", "The password to log into the controller with (you can use environment variables)")
	flag.StringVar(&unifiRuleNumber, "rule", "2000", "The firewal rule number in LAN_IN to toggle")
	flag.Parse()

	fmt.Printf("Logging into unifi system at %s with user %s to toggle rule %s\n", unifiIp, unifiUsername, unifiRuleNumber)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1920, 2080),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var buf []byte
	if err = chromedp.Run(ctx, ToggleFirewallRule(unifiIp, unifiUsername, unifiPassword, unifiRuleNumber)); err != nil {
		fmt.Printf("Error encountered: %s\n", err.Error())
		os.Exit(1)
	}

	if err = chromedp.Run(ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		fmt.Printf("Error taking screenshot: %s\n", err.Error())
		os.Exit(1)
	}

	if err := ioutil.WriteFile("end.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Finished!")

}

func ToggleFirewallRule(ip string, username string, password string, rule string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(fmt.Sprintf("https://%s", ip)),
		chromedp.WaitVisible(ButtonWithText("Sign In")),
		chromedp.SendKeys("#login-username", username, chromedp.ByID, chromedp.NodeVisible),
		chromedp.SendKeys("#login-password", password, chromedp.ByID, chromedp.NodeVisible),
		chromedp.Click(ButtonWithText("Sign In")),
		chromedp.WaitVisible("//span[text()='Network']"),
		chromedp.Navigate(fmt.Sprintf("https://%s/network/site/default/settings/firewall/rules/ipv4/list/LAN_IN", ip)),
		chromedp.Click(fmt.Sprintf("//td[contains(@class,'firewallRulesIndex') and text()='%s']/../td[contains(@class,'firewallRulesActions')]/button[contains(.,'Edit')]", rule), chromedp.NodeVisible),
		chromedp.Click("//input[@id='firewallRuleEnabled']", chromedp.NodeVisible),
		chromedp.Sleep(100 * time.Millisecond),
		chromedp.Click(ButtonWithText("Save"), chromedp.NodeVisible),
		chromedp.Sleep(100 * time.Millisecond),
	}
}

func ButtonWithText(text string) string {
	return fmt.Sprintf("//button[contains(.,'%s')]", text)
}
