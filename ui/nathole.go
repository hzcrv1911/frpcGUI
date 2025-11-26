package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/hzcrv1911/frpcgui/i18n"
	"github.com/hzcrv1911/frpcgui/pkg/res"
)

type NATDiscoveryDialog struct {
	*walk.Dialog

	table   *walk.TableView
	barView *walk.ProgressBar

	// STUN server address
	serverAddr string
	closed     bool
}

func NewNATDiscoveryDialog(serverAddr string) *NATDiscoveryDialog {
	return &NATDiscoveryDialog{serverAddr: serverAddr}
}

func (nd *NATDiscoveryDialog) Run(owner walk.Form) (int, error) {
	dlg := NewBasicDialog(&nd.Dialog, i18n.Sprintf("NAT Discovery"), loadIcon(res.IconNat, 32),
		DataBinder{}, nil,
		VSpacer{Size: 1},
		Composite{
			Layout: HBox{MarginsZero: true},
			Children: []Widget{
				Label{Text: i18n.SprintfColon("STUN Server")},
				TextEdit{Text: nd.serverAddr, ReadOnly: true, CompactHeight: true},
			},
		},
		VSpacer{Size: 1},
		TableView{
			Name:     "tb",
			Visible:  false,
			AssignTo: &nd.table,
			Columns: []TableViewColumn{
				{Title: i18n.Sprintf("Item"), DataMember: "Title", Width: 180},
				{Title: i18n.Sprintf("Value"), DataMember: "Value", Width: 180},
			},
			ColumnsOrderable: false,
		},
		ProgressBar{AssignTo: &nd.barView, Visible: Bind("!tb.Visible"), MarqueeMode: true},
		VSpacer{},
	)
	dlg.MinSize = Size{Width: 400, Height: 350}
	if err := dlg.Create(owner); err != nil {
		return 0, err
	}
	nd.barView.SetFocus()
	nd.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		nd.closed = true
	})

	// Start discovering NAT type
	// NAT discovery is not available when using WinSW integration
	go func() {
		nd.Synchronize(func() {
			if !nd.closed {
				nd.barView.SetMarqueeMode(false)
				showErrorMessage(nd.Form(), "", i18n.Sprintf("NAT discovery is not available when using WinSW integration."))
				nd.Cancel()
			}
		})
	}()

	return nd.Dialog.Run(), nil
}
