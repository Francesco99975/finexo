// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.857
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/views/components"
	"github.com/Francesco99975/finexo/views/icons"
	"github.com/Francesco99975/finexo/views/layouts"
)

func Index(site models.Site, csrf, nonce string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<main class=\"flex-grow container mx-auto px-4 py-8 max-w-4xl transition-colors\"><h1 class=\"text-3xl font-bold text-text-primary mb-6 text-center\">Investment Compound Calculator</h1>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = components.SearchBar().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<form id=\"calculate-form\" class=\"space-y-4 mb-8 relative\" hx-post=\"/calculate\" hx-target=\"#calculation-results\" hx-indicator=\"#calculate-indicator\"><input type=\"hidden\" name=\"_csrf\" id=\"_csrf\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(csrf)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `index.templ`, Line: 22, Col: 61}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "\"><!-- Security Information Display --><div id=\"security-info-indicator\" class=\"htmx-indicator absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center z-10 pointer-events-none\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = icons.SelectedLoading().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "</div><!-- Initial state - No security selected --><div class=\"bg-bg-std rounded-lg shadow-md p-6 border-l-4 border-l-primary\"><h2 class=\"text-lg font-semibold text-text-primary mb-4\">No Security Selected</h2><p class=\"text-text-secondary mb-4\">Search and select a security above or enter a custom interest rate below.</p><div class=\"mb-4\"><label for=\"rate\" class=\"block text-sm font-medium text-text-primary mb-1\">HISA Interest Rate (%)</label> <input type=\"number\" id=\"rate\" name=\"rate\" class=\"block w-full p-2 border border-std rounded-md focus:ring-accent focus:border-accent\" value=\"1.5\" min=\"0\" step=\"0.1\"></div><!-- Compounding Frequency --><div><label for=\"compoundingfrequency\" class=\"block text-sm font-medium text-text-primary mb-1\">Compounding Frequency</label> <select id=\"compoundingfrequency\" name=\"compoundingfrequency\" class=\"block w-full p-2 border border-std rounded-md focus:ring-accent focus:border-accent\"><option value=\"daily\">Daily</option> <option value=\"weekly\">Weekly</option> <option value=\"monthly\" selected>Monthly</option> <option value=\"quarterly\">Quarterly</option> <option value=\"semi-annually\">Semi-Annually</option> <option value=\"annually\">Annually</option></select></div><!-- Currency Selection --><div class=\"mb-4\"><label for=\"currency\" class=\"block text-sm font-medium text-text-primary mb-1\">Currency</label> <select id=\"currency\" name=\"currency\" class=\"block w-full p-2 border border-std rounded-md focus:ring-accent focus:border-accent\"><option value=\"CAD\" selected>CAD</option> <option value=\"EUR\">EUR</option> <option value=\"USD\">USD</option></select></div></div><!-- Calculator Form --><div class=\"bg-bg-std rounded-lg shadow-md p-6 border-l-4 border-l-accent\" x-data=\"{ contributionFrequency: &#39;monthly&#39; }\"><h2 class=\"text-lg font-semibold text-text-primary mb-4\">Compound Calculator</h2><input type=\"hidden\" name=\"sid\" id=\"sid\" value=\"default\"><!-- Principal Amount --><div><label for=\"principal\" class=\"block text-sm font-medium text-text-primary mb-1\">Initial Investment</label><div class=\"relative\"><div class=\"absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none\"><span class=\"text-text-secondary\">$</span></div><input type=\"number\" id=\"principal\" name=\"principal\" class=\"block w-full pl-8 p-2 border border-std rounded-md focus:ring-accent focus:border-accent\" value=\"10000\" min=\"0\"></div></div><!-- Contribution Frequency --><div><label for=\"contribfrequency\" class=\"block text-sm font-medium text-text-primary mb-1\">Contribution Frequency</label> <select id=\"contribfrequency\" name=\"contribfrequency\" class=\"block w-full p-2 border border-std rounded-md focus:ring-accent focus:border-accent\" x-model=\"contributionFrequency\"><option value=\"weekly\">Weekly</option> <option value=\"monthly\">Monthly</option> <option value=\"quarterly\">Quarterly</option> <option value=\"semi-annually\">Semi-Annually</option> <option value=\"annually\">Annually</option></select></div><!-- Contribution Amount --><div><label for=\"contribution\" class=\"block text-sm font-medium text-text-primary mb-1\"><span x-text=\"contributionFrequency.charAt(0).toUpperCase() + contributionFrequency.slice(1)\">Monthly</span> Contribution</label><div class=\"relative\"><div class=\"absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none\"><span class=\"text-text-secondary\">$</span></div><input type=\"number\" id=\"contribution\" name=\"contribution\" class=\"block w-full pl-8 p-2 border border-std rounded-md focus:ring-accent focus:border-accent\" value=\"1000\" min=\"0\"></div></div><!-- Compounding Years --><div><label for=\"years\" class=\"block text-sm font-medium text-text-primary mb-1\">Compounding Years</label> <input type=\"number\" id=\"years\" name=\"years\" class=\"block w-full p-2 border border-std rounded-md focus:ring-accent focus:border-accent\" value=\"10\" min=\"1\" max=\"50\"></div><!-- Submit Button --><div class=\"relative mt-10\"><button type=\"submit\" class=\"w-full bg-accent text-white py-2 px-4 rounded-md hover:bg-accent/90 focus:outline-none focus:ring-2 focus:ring-accent/50 focus:ring-offset-2 transition-colors\">Calculate Results</button><div id=\"calculate-indicator\" class=\"htmx-indicator absolute inset-0 flex items-center justify-center bg-accent bg-opacity-75 rounded-md pointer-events-none\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = icons.CalculateLoading().Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "</div></div></div></form><!-- Calculation Results (will be populated by server) --><div id=\"calculation-results\" class=\"mt-8\"><!-- Results will be inserted here by the server --></div></main>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = layouts.CoreHTML(site, nonce, nil, nil, nil).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
