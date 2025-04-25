package tools

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// GeneratePDF generates a PDF file with investment calculation results
func GeneratePDF(results helpers.CalculationResults) (string, error) {
	// Generate filename with timestamp
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s-%d.pdf", results.SID, timestamp)

	contributionsMonthOneStr := helpers.NormalizeFloatStrToIntStr(strings.ReplaceAll(results.YearResults[0].MonthsResults[0].Contributions[3:], " ", ""))

	contributionsMonthOneInt, err := strconv.Atoi(contributionsMonthOneStr)
	if err != nil {
		return "", err
	}

	PrincipalStr := helpers.NormalizeFloatStrToIntStr(strings.ReplaceAll(results.Principal[3:], " ", ""))
	PrincipalInt, err := strconv.Atoi(PrincipalStr)
	if err != nil {
		return "", err
	}

	contrib := float64((contributionsMonthOneInt - PrincipalInt) / 100)
	contribStr, err := helpers.FormatPrice(contrib, results.Currency)
	if err != nil {
		return "", err
	}

	// Configure PDF settings (A4, horizontal orientation)
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithOrientation(orientation.Horizontal).
		Build()
	m := maroto.New(cfg)

	// Define color constants
	darkBlue := &props.Color{Red: 0, Green: 51, Blue: 102}     // Dark blue for headers
	lightBlue := &props.Color{Red: 0, Green: 102, Blue: 204}   // Light blue for highlights
	gray := &props.Color{Red: 50, Green: 50, Blue: 50}         // Gray for subtitles
	white := &props.Color{Red: 255, Green: 255, Blue: 255}     // White for backgrounds
	lightGray := &props.Color{Red: 240, Green: 240, Blue: 240} // Light gray for alternating rows
	black := &props.Color{Red: 0, Green: 0, Blue: 0}           // Black for borders

	var title string
	if results.Price != "" {
		title = fmt.Sprintf("Investment Summary for %s with a starting price of %s", results.SID, results.Price)
	} else {
		title = "HISA Calculation"
	}

	// Add Title
	m.AddRows(
		row.New(25).Add(
			col.New(12).Add(
				text.New(title, props.Text{
					Size:  18,
					Align: align.Center,
					Style: fontstyle.Bold,
					Color: darkBlue,
				}),
			),
		),
	)

	// Add Subtitle
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Calculated compounding growth for %d years with a principal of %s @ a starting rate of %s", len(results.YearResults), results.Principal, results.Rate), props.Text{
					Size:  12,
					Align: align.Center,
					Style: fontstyle.Italic,
					Color: gray,
				}),
			),
		),
	)

	// Overview Section Header
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New("Overview", props.Text{
					Size:  14,
					Style: fontstyle.Bold,
					Color: lightBlue,
				}),
			),
		),
	)

	// Overview Data
	overviewData := [][]string{
		{"SID", results.SID},
		{"Principal", results.Principal},
		{"Rate", results.Rate},
		{"Rate Frequency", results.RateFreq},
		{"Currency", results.Currency},
		{"Profit", results.Profit},
		{"Total Contributions", results.TotalContributions},
		{"Contribution Frequency", results.ContribFreq},
		{"Final Balance", results.FinalBalance},
		{"Contribution", contribStr},
	}
	m.AddRows(buildOverviewTable(overviewData, lightGray, lightBlue, black)...)

	// Add spacing
	m.AddRows(row.New(15))

	// Yearly Breakdown
	for _, year := range results.YearResults {
		m.AddRows(row.New(0)) // Page break
		m.AddRows(
			row.New(12).Add(
				col.New(12).Add(
					text.New(year.YearName, props.Text{
						Size:  14,
						Style: fontstyle.Bold,
						Color: lightBlue,
					}),
				),
			),
		)
		m.AddRows(buildKeyValueTable([][]string{
			{"Share Amount", year.ShareAmount},
			{"Total Year Gains", year.TotalYearGains},
			{"Cumulative Gain", year.CumGain},
			{"YoY Growth", year.YoyGrowth},
			{"Total Growth", year.TotalGrowth},
			{"Balance", year.Balance},
		}, 10, lightGray, black)...)
		m.AddRows(
			row.New(8).Add(
				col.New(12).Add(
					text.New("Monthly Results", props.Text{
						Size:  12,
						Style: fontstyle.BoldItalic,
						Color: gray,
					}),
				),
			),
		)
		m.AddRows(buildMonthTable(year.MonthsResults, white, lightGray, darkBlue, black)...)
	}

	// Generate and save the PDF
	doc, err := m.Generate()
	if err != nil {
		return "", err
	}
	err = doc.Save(filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}

// buildOverviewTable creates a two-column grid for overview data with borders
func buildOverviewTable(data [][]string, labelBgColor *props.Color, highlightColor *props.Color, borderColor *props.Color) []core.Row {
	var rows []core.Row
	for i := 0; i < len(data); i += 2 {
		r := row.New(10)
		// First pair
		key1 := data[i][0]
		value1 := data[i][1]
		valueProps := props.Text{Size: 10, Align: align.Center}
		if key1 == "Final Balance" || key1 == "Profit" {
			valueProps.Style = fontstyle.Bold
			valueProps.Color = highlightColor
		}
		r.Add(
			col.New(2).Add(
				text.New(key1+":", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
					Align: align.Right,
				}),
			).WithStyle(&props.Cell{
				BackgroundColor: labelBgColor,
				BorderType:      border.Full,
				BorderThickness: 0.2,
				BorderColor:     borderColor,
			}),
			col.New(4).Add(
				text.New(value1, valueProps),
			).WithStyle(&props.Cell{
				BorderType:      border.Full,
				BorderThickness: 0.2,
				BorderColor:     borderColor,
			}),
		)
		// Second pair, if exists
		if i+1 < len(data) {
			key2 := data[i+1][0]
			value2 := data[i+1][1]
			valueProps2 := props.Text{Size: 10, Align: align.Center}
			if key2 == "Final Balance" || key2 == "Profit" {
				valueProps2.Style = fontstyle.Bold
				valueProps2.Color = highlightColor
			}
			r.Add(
				col.New(2).Add(
					text.New(key2+":", props.Text{
						Size:  10,
						Style: fontstyle.Bold,
						Align: align.Right,
					}),
				).WithStyle(&props.Cell{
					BackgroundColor: labelBgColor,
					BorderType:      border.Full,
					BorderThickness: 0.2,
					BorderColor:     borderColor,
				}),
				col.New(4).Add(
					text.New(value2, valueProps2),
				).WithStyle(&props.Cell{
					BorderType:      border.Full,
					BorderThickness: 0.2,
					BorderColor:     borderColor,
				}),
			)
		} else {
			// Add empty columns with borders for odd number of pairs
			r.Add(
				col.New(2).Add(
					text.New("", props.Text{}),
				).WithStyle(&props.Cell{
					BorderType:      border.Full,
					BorderThickness: 0.2,
					BorderColor:     borderColor,
				}),
				col.New(4).Add(
					text.New("", props.Text{}),
				).WithStyle(&props.Cell{
					BorderType:      border.Full,
					BorderThickness: 0.2,
					BorderColor:     borderColor,
				}),
			)
		}
		rows = append(rows, r)
	}
	return rows
}

// buildKeyValueTable creates a key-value table with borders
func buildKeyValueTable(data [][]string, rowHeight float64, labelBgColor *props.Color, borderColor *props.Color) []core.Row {
	var rows []core.Row
	for _, pair := range data {
		labelCol := col.New(4).Add(
			text.New(pair[0], props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		).WithStyle(&props.Cell{
			BackgroundColor: labelBgColor,
			BorderType:      border.Full,
			BorderThickness: 0.2,
			BorderColor:     borderColor,
		})
		valueCol := col.New(8).Add(
			text.New(pair[1], props.Text{
				Size:  10,
				Align: align.Center,
			}),
		).WithStyle(&props.Cell{
			BorderType:      border.Full,
			BorderThickness: 0.2,
			BorderColor:     borderColor,
		})
		rows = append(rows, row.New(rowHeight).Add(labelCol, valueCol))
	}
	return rows
}

// buildMonthTable creates a table for monthly results with borders
func buildMonthTable(months []helpers.MonthCalcResults, white *props.Color, lightGray *props.Color, darkBlue *props.Color, borderColor *props.Color) []core.Row {
	// Define headers
	headers := []string{"Month", "Shares", "Contrib.", "Price Gain", "Div. Gain", "Monthly Gain", "Cum. Gain", "Balance", "Return", "DRIP"}
	headerRow := row.New(12).Add(
		col.New(2).Add(text.New(headers[0], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[1], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[2], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[3], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[4], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[5], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[6], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(2).Add(text.New(headers[7], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[8], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		col.New(1).Add(text.New(headers[9], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: white})).WithStyle(&props.Cell{BackgroundColor: darkBlue, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
	)
	rows := []core.Row{headerRow}

	// Add data rows with alternating background colors
	for i, month := range months {
		bgColor := white
		if i%2 == 1 {
			bgColor = lightGray
		}
		rows = append(rows, row.New(9).Add(
			col.New(2).Add(text.New(month.MonthName, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.ShareAmount, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.Contributions, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.MonthlyGainedFromPriceInc, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.MonthlyGainedFromDividends, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.MonthlyGain, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.CumGain, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(2).Add(text.New(month.Balance, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.Return, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
			col.New(1).Add(text.New(month.DRIP, props.Text{Size: 9, Align: align.Center})).WithStyle(&props.Cell{BackgroundColor: bgColor, BorderType: border.Full, BorderThickness: 0.2, BorderColor: borderColor}),
		))
	}
	return rows
}
