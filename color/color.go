package color

import "github.com/fatih/color"

var ErrColor = color.New(color.FgHiRed).Add(color.Bold)
var PlanColor = color.New(color.FgHiGreen).Add(color.Bold)
var ResourcePlanColor = color.New(color.FgHiBlue).Add(color.Bold)
var NoColor = color.New()