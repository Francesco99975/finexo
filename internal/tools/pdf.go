package tools

import (
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func GeneratePDF(results helpers.CalculationResults) (string, error) {
	filename := "investment_summary.pdf"
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithOrientation(orientation.Horizontal).
		Build()
	m := maroto.New(cfg)

	// Title
	m.AddRows(
		row.New(25).Add(
			col.New(12).Add(
				text.New("Investment Summary", props.Text{
					Size:  18,
					Align: align.Center,
					Style: fontstyle.Bold,
					Color: &props.Color{Red: 0, Green: 51, Blue: 102},
				}),
			),
		),
		row.New(10), // Spacing
	)

	// Overview Section
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New("Overview", props.Text{
					Size:  14,
					Style: fontstyle.Bold,
					Color: &props.Color{Red: 0, Green: 102, Blue: 204},
				}),
			),
		),
	)
	m.AddRows(buildKeyValueTable([][]string{
		{"SID", results.SID},
		{"Principal", results.Principal},
		{"Rate", results.Rate},
		{"Rate Frequency", results.RateFreq},
		{"Currency", results.Currency},
		{"Profit", results.Profit},
		{"Total Contributions", results.TotalContributions},
		{"Contribution Frequency", results.ContribFreq},
		{"Final Balance", results.FinalBalance},
	}, 10, &props.Color{Red: 240, Green: 240, Blue: 240})...)
	m.AddRows(row.New(15)) // Spacing

	// Yearly Breakdown
	for _, year := range results.YearResults {
		m.AddRows(row.New(0)) // Page break
		m.AddRows(
			row.New(12).Add(
				col.New(12).Add(
					text.New(year.YearName, props.Text{
						Size:  14,
						Style: fontstyle.Bold,
						Color: &props.Color{Red: 0, Green: 102, Blue: 204},
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
		}, 10, &props.Color{Red: 240, Green: 240, Blue: 240})...)
		m.AddRows(
			row.New(8).Add(
				col.New(12).Add(
					text.New("Monthly Results", props.Text{
						Size:  12,
						Style: fontstyle.BoldItalic,
						Color: &props.Color{Red: 50, Green: 50, Blue: 50},
					}),
				),
			),
		)
		m.AddRows(buildMonthTable(year.MonthsResults)...)
	}

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

func buildKeyValueTable(data [][]string, rowHeight float64, labelBgColor *props.Color) []core.Row {
	var rows []core.Row
	for _, pair := range data {
		labelCol := col.New(4).Add(
			text.New(pair[0], props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Left,
			}),
		)
		if labelBgColor != nil {
			labelCol = labelCol.WithStyle(&props.Cell{BackgroundColor: labelBgColor})
		}
		rows = append(rows, row.New(rowHeight).Add(
			labelCol,
			col.New(8).Add(
				text.New(pair[1], props.Text{
					Size:  10,
					Align: align.Left,
				}),
			),
		))
	}
	return rows
}

func buildMonthTable(months []helpers.MonthCalcResults) []core.Row {
	headers := []string{"Month", "Shares", "Contrib.", "Price Gain", "Div. Gain", "Monthly Gain", "Cum. Gain", "Balance", "Return", "DRIP"}
	headerRow := row.New(12).Add(
		col.New(2).Add(text.New(headers[0], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[1], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[2], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[3], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[4], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[5], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[6], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(2).Add(text.New(headers[7], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[8], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
		col.New(1).Add(text.New(headers[9], props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.Color{Red: 255, Green: 255, Blue: 255}})).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 80, Green: 80, Blue: 80}}),
	)
	rows := []core.Row{headerRow}
	for i, month := range months {
		bgColor := &props.Color{Red: 255, Green: 255, Blue: 255} // White
		if i%2 == 1 {
			bgColor = &props.Color{Red: 240, Green: 240, Blue: 240} // Light gray
		}
		rows = append(rows, row.New(9).WithStyle(&props.Cell{BackgroundColor: bgColor}).Add(
			col.New(2).Add(text.New(month.MonthName, props.Text{Size: 9, Align: align.Left})),
			col.New(1).Add(text.New(month.ShareAmount, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.Contributions, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.MonthlyGainedFromPriceInc, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.MonthlyGainedFromDividends, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.MonthlyGain, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.CumGain, props.Text{Size: 9, Align: align.Right})),
			col.New(2).Add(text.New(month.Balance, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.Return, props.Text{Size: 9, Align: align.Right})),
			col.New(1).Add(text.New(month.DRIP, props.Text{Size: 9, Align: align.Right})),
		))
	}
	return rows
}
