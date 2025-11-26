package ui

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/hzcrv1911/frpcgui/i18n"
	"github.com/hzcrv1911/frpcgui/pkg/consts"
	"github.com/hzcrv1911/frpcgui/pkg/res"
	"github.com/hzcrv1911/frpcgui/pkg/util"
	"github.com/hzcrv1911/frpcgui/services"
)

var configStateDescription = map[consts.ConfigState]string{
	consts.ConfigStateUnknown:      i18n.Sprintf("Not Installed"),
	consts.ConfigStateStarted:      i18n.Sprintf("Started"),
	consts.ConfigStateStopped:      i18n.Sprintf("Stopped"),
	consts.ConfigStateStarting:     i18n.Sprintf("Starting"),
	consts.ConfigStateStopping:     i18n.Sprintf("Stopping"),
	consts.ConfigStateNotInstalled: i18n.Sprintf("Not Installed"),
}

type PanelView struct {
	*walk.GroupBox

	stateText   *walk.Label
	stateImage  *walk.ImageView
	addressText *walk.Label
	protoText   *walk.Label
	protoImage  *walk.ImageView
	toggleBtn   *walk.PushButton
	serviceBtn  *walk.PushButton
}

func NewPanelView() *PanelView {
	return new(PanelView)
}

func (pv *PanelView) View() Widget {
	var cpIcon *walk.CustomWidget
	cpIconColor := res.ColorDarkGray
	setCopyIconColor := func(button walk.MouseButton, color walk.Color) {
		if button == walk.LeftButton {
			cpIconColor = color
			cpIcon.Invalidate()
		}
	}
	return GroupBox{
		AssignTo: &pv.GroupBox,
		Title:    "",
		Layout:   Grid{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 10},
		Children: []Widget{
			Label{Text: i18n.SprintfColon("Status"), Row: 0, Column: 0, Alignment: AlignHFarVCenter},
			Label{Text: i18n.SprintfColon("Server Address"), Row: 1, Column: 0, Alignment: AlignHFarVCenter},
			Label{Text: i18n.SprintfColon("Protocol"), Row: 2, Column: 0, Alignment: AlignHFarVCenter},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    0, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					ImageView{AssignTo: &pv.stateImage, Margin: 0},
					HSpacer{Size: 4},
					Label{AssignTo: &pv.stateText},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true, MarginsZero: true},
				Row:    1, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					Label{AssignTo: &pv.addressText},
					HSpacer{Size: 5},
					CustomWidget{
						AssignTo:            &cpIcon,
						Background:          TransparentBrush{},
						ClearsBackground:    true,
						InvalidatesOnResize: true,
						MinSize:             Size{Width: 16, Height: 16},
						ToolTipText:         i18n.Sprintf("Copy"),
						PaintPixels: func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
							return drawCopyIcon(canvas, cpIconColor)
						},
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							setCopyIconColor(button, res.ColorLightBlue)
						},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							setCopyIconColor(button, res.ColorDarkGray)
							bounds := cpIcon.ClientBoundsPixels()
							if x >= 0 && x <= bounds.Right() && y >= 0 && y <= bounds.Bottom() {
								walk.Clipboard().SetText(pv.addressText.Text())
							}
						},
					},
					VSpacer{Size: 20},
				},
			},
			Composite{
				Layout: HBox{Spacing: 2, MarginsZero: true},
				Row:    2, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					ImageView{
						AssignTo:    &pv.protoImage,
						Image:       loadIcon(res.IconFlatLock, 14),
						ToolTipText: i18n.Sprintf("Your connection to the server is encrypted"),
					},
					Label{AssignTo: &pv.protoText},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 5},
				Row:    3, Column: 1,
				Alignment: AlignHNearVCenter,
				Children: []Widget{
					PushButton{
						AssignTo:  &pv.toggleBtn,
						Text:      i18n.Sprintf("Start"),
						MaxSize:   Size{Width: 80},
						Enabled:   false,
						OnClicked: pv.ToggleService,
					},
					PushButton{
						AssignTo:  &pv.serviceBtn,
						Text:      i18n.Sprintf("Install Service"),
						MaxSize:   Size{Width: 140},
						OnClicked: pv.toggleServiceInstallation,
					},
					HSpacer{},
				},
			},
		},
	}
}

func (pv *PanelView) OnCreate() {
	pv.setState(consts.ConfigStateUnknown)
}

func (pv *PanelView) setState(state consts.ConfigState) {
	pv.stateImage.SetImage(iconForConfigState(state, 14))
	pv.stateText.SetText(configStateDescription[state])
	// Disable button when starting, stopping, unknown, or not installed
	shouldEnable := true
	switch state {
	case consts.ConfigStateStarting, consts.ConfigStateStopping,
		consts.ConfigStateUnknown, consts.ConfigStateNotInstalled:
		shouldEnable = false
	}
	pv.toggleBtn.SetEnabled(shouldEnable)

	switch state {
	case consts.ConfigStateStarted, consts.ConfigStateStopping:
		pv.toggleBtn.SetText(i18n.Sprintf("Stop"))
	default:
		pv.toggleBtn.SetText(i18n.Sprintf("Start"))
	}
	pv.updateServiceButton(state)
}

func (pv *PanelView) ToggleService() {
	conf := getCurrentConf()
	if conf == nil {
		return
	}
	var err error
	if conf.State == consts.ConfigStateStarted {
		if walk.MsgBox(pv.Form(), i18n.Sprintf("Stop config \"%s\"", conf.Name()),
			i18n.Sprintf("Are you sure you would like to stop config \"%s\"?", conf.Name()),
			walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdNo {
			return
		}
		err = pv.StopServiceOnly(conf)
	} else {
		if !util.FileExists(conf.Path) {
			warnConfigRemoved(pv.Form(), conf.Name())
			return
		}
		err = pv.StartServiceOnly(conf)
	}
	if err != nil {
		showError(err, pv.Form())
	}
}

func (pv *PanelView) toggleServiceInstallation() {
	conf := getCurrentConf()
	if conf == nil {
		return
	}
	switch conf.State {
	case consts.ConfigStateNotInstalled, consts.ConfigStateUnknown:
		pv.InstallServiceOnly()
	default:
		pv.UninstallServiceOnly()
	}
}

func (pv *PanelView) updateServiceButton(state consts.ConfigState) {
	if pv.serviceBtn == nil {
		return
	}
	current := getCurrentConf()
	if current == nil {
		pv.serviceBtn.SetText(i18n.Sprintf("Install Service"))
		pv.serviceBtn.SetEnabled(false)
		return
	}
	isInstallState := state == consts.ConfigStateNotInstalled || state == consts.ConfigStateUnknown
	if isInstallState {
		pv.serviceBtn.SetText(i18n.Sprintf("Install Service"))
	} else {
		pv.serviceBtn.SetText(i18n.Sprintf("Uninstall Service"))
	}
	shouldEnable := state != consts.ConfigStateStarting && state != consts.ConfigStateStopping
	pv.serviceBtn.SetEnabled(shouldEnable)
}

// InstallServiceOnly installs the WinSW service without starting it
func (pv *PanelView) InstallServiceOnly() {
	conf := getCurrentConf()
	if conf == nil {
		return
	}
	if !util.FileExists(conf.Path) {
		warnConfigRemoved(pv.Form(), conf.Name())
		return
	}
	// Ensure log directory is valid
	if logFile := conf.Data.LogFile; logFile != "" && logFile != "console" {
		if err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm); err != nil {
			showError(err, pv.Form())
			return
		}
	}
	go func(conf *Conf) {
		if err := services.InstallWinSWService(conf.Name(), conf.Path, !conf.Data.AutoStart()); err != nil {
			pv.Synchronize(func() {
				showErrorMessage(pv.Form(), i18n.Sprintf("Install service for config \"%s\"", conf.Name()), err.Error())
			})
		} else {
			pv.Synchronize(func() {
				setConfState(conf, consts.ConfigStateStopped)
				if getCurrentConf() == conf {
					pv.setState(consts.ConfigStateStopped)
				}
				walk.MsgBox(pv.Form(), i18n.Sprintf("Success"),
					i18n.Sprintf("Service installed successfully for config \"%s\"", conf.Name()),
					walk.MsgBoxOK|walk.MsgBoxIconInformation)
			})
		}
	}(conf)
}

// UninstallServiceOnly uninstalls the WinSW service
func (pv *PanelView) UninstallServiceOnly() {
	conf := getCurrentConf()
	if conf == nil {
		return
	}
	if walk.MsgBox(pv.Form(), i18n.Sprintf("Uninstall service for \"%s\"", conf.Name()),
		i18n.Sprintf("Are you sure you would like to uninstall the service for config \"%s\"?", conf.Name()),
		walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdNo {
		return
	}
	go func(conf *Conf) {
		if err := services.UninstallService(conf.Path, false); err != nil {
			pv.Synchronize(func() {
				showErrorMessage(pv.Form(), i18n.Sprintf("Uninstall service for config \"%s\"", conf.Name()), err.Error())
			})
		} else {
			pv.Synchronize(func() {
				setConfState(conf, consts.ConfigStateNotInstalled)
				if getCurrentConf() == conf {
					pv.setState(consts.ConfigStateNotInstalled)
				}
				walk.MsgBox(pv.Form(), i18n.Sprintf("Success"),
					i18n.Sprintf("Service uninstalled successfully for config \"%s\"", conf.Name()),
					walk.MsgBoxOK|walk.MsgBoxIconInformation)
			})
		}
	}(conf)
}

// StartServiceOnly starts an already installed service
func (pv *PanelView) StartServiceOnly(conf *Conf) error {
	if err := services.VerifyClientConfig(conf.Path); err != nil {
		return err
	}
	oldState := conf.State
	setConfState(conf, consts.ConfigStateStarting)
	pv.setState(consts.ConfigStateStarting)
	go func() {
		if err := services.StartWinSWService(conf.Path); err != nil {
			pv.Synchronize(func() {
				showErrorMessage(pv.Form(), i18n.Sprintf("Start config \"%s\"", conf.Name()), err.Error())
				if conf.State == consts.ConfigStateStarting {
					setConfState(conf, oldState)
					if getCurrentConf() == conf {
						pv.setState(oldState)
					}
				}
			})
		}
	}()
	return nil
}

// StopServiceOnly stops a running service without uninstalling it
func (pv *PanelView) StopServiceOnly(conf *Conf) (err error) {
	oldState := conf.State
	setConfState(conf, consts.ConfigStateStopping)
	pv.setState(consts.ConfigStateStopping)
	defer func() {
		if err != nil {
			setConfState(conf, oldState)
			pv.setState(oldState)
		}
	}()
	err = services.StopWinSWService(conf.Path)
	return
}

// StartService creates a daemon service of the given config, then starts it (deprecated, use InstallServiceOnly + StartServiceOnly)
func (pv *PanelView) StartService(conf *Conf) error {
	// Verify the config file
	if err := services.VerifyClientConfig(conf.Path); err != nil {
		return err
	}
	// Ensure log directory is valid
	if logFile := conf.Data.LogFile; logFile != "" && logFile != "console" {
		if err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm); err != nil {
			return err
		}
	}
	oldState := conf.State
	setConfState(conf, consts.ConfigStateStarting)
	pv.setState(consts.ConfigStateStarting)
	go func() {
		if err := services.InstallService(conf.Name(), conf.Path, !conf.Data.AutoStart()); err != nil {
			pv.Synchronize(func() {
				showErrorMessage(pv.Form(), i18n.Sprintf("Start config \"%s\"", conf.Name()), err.Error())
				if conf.State == consts.ConfigStateStarting {
					setConfState(conf, oldState)
					if getCurrentConf() == conf {
						pv.setState(oldState)
					}
				}
			})
		}
	}()
	return nil
}

// StopService stops the service of the given config, then removes it (deprecated, use StopServiceOnly + UninstallServiceOnly)
func (pv *PanelView) StopService(conf *Conf) (err error) {
	oldState := conf.State
	setConfState(conf, consts.ConfigStateStopping)
	pv.setState(consts.ConfigStateStopping)
	defer func() {
		if err != nil {
			setConfState(conf, oldState)
			pv.setState(oldState)
		}
	}()
	err = services.UninstallService(conf.Path, false)
	return
}

// Invalidate updates views using the current config
func (pv *PanelView) Invalidate(state bool) {
	conf := getCurrentConf()
	if conf == nil {
		pv.SetTitle("")
		pv.setState(consts.ConfigStateUnknown)
		pv.addressText.SetText("")
		pv.protoText.SetText("")
		pv.protoImage.SetVisible(false)
		return
	}
	data := conf.Data
	if pv.Title() != conf.Name() {
		pv.SetTitle(conf.Name())
	}
	addr := data.ServerAddress
	if addr == "" {
		addr = "0.0.0.0"
	}
	if pv.addressText.Text() != addr {
		pv.addressText.SetText(addr)
	}
	pv.protoImage.SetVisible(data.TLSEnable || data.Protocol == consts.ProtoWSS || data.Protocol == consts.ProtoQUIC)
	proto := util.GetOrElse(data.Protocol, consts.ProtoTCP)
	if proto == consts.ProtoWebsocket {
		proto = "ws"
	}
	proto = strings.ToUpper(proto)
	if data.HTTPProxy != "" && data.Protocol != consts.ProtoQUIC {
		if u, err := url.Parse(data.HTTPProxy); err == nil {
			proto += " + " + strings.ToUpper(u.Scheme)
		}
	}
	if pv.protoText.Text() != proto {
		pv.protoText.SetText(proto)
	}
	if state {
		pv.setState(conf.State)
	}
}
