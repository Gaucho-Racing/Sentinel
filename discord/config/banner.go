package config

import "github.com/fatih/color"

var Banner = `
███████╗███████╗███╗   ██╗████████╗██╗███╗   ██╗███████╗██╗
██╔════╝██╔════╝████╗  ██║╚══██╔══╝██║████╗  ██║██╔════╝██║
███████╗█████╗  ██╔██╗ ██║   ██║   ██║██╔██╗ ██║█████╗  ██║
╚════██║██╔══╝  ██║╚██╗██║   ██║   ██║██║╚██╗██║██╔══╝  ██║
███████║███████╗██║ ╚████║   ██║   ██║██║ ╚████║███████╗███████╗
╚══════╝╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝
`

func PrintStartupBanner() {
	banner := color.New(color.Bold, color.FgHiCyan).PrintlnFunc()
	banner(Banner)
	version := color.New(color.Bold, color.FgCyan).PrintlnFunc()
	version("Running " + Service.FormattedNameWithVersion() + " [ENV: " + Env + "]")
	println()
}
