package helper

import (
	"strings"
)

func FormatKeyLogs(logs string) string {
	// Replace all instances of "L_SHIFT", "R_SHIFT", "L_CTRL", "R_CTRL" with ""
	logs = strings.Replace(logs, "L_SHIFTL_SHIFT", " <1_LS> ", -1)
	logs = strings.Replace(logs, "R_SHIFTR_SHIFT", " <1_RS> ", -1)
	logs = strings.Replace(logs, "L_CTRLL_CTRL", " <1_LC> ", -1)
	logs = strings.Replace(logs, "R_CTRLR_CTRL", " <1_RC> ", -1)
	logs = strings.Replace(logs, "L_ALTL_ALT", " <1_L_ALT> ", -1)
	logs = strings.Replace(logs, "R_ALTR_ALT", " <1_R_ALT> ", -1)
	logs = strings.Replace(logs, "L_SHIFT", " <H_LS> ", -1)
	logs = strings.Replace(logs, "R_SHIFT", " <H_RS> ", -1)
	logs = strings.Replace(logs, "L_CTRL", " <1_LC> ", -1)
	logs = strings.Replace(logs, "R_CTRL", " <1_RC> ", -1)
	logs = strings.Replace(logs, "SPACESPACE", " ", -1)
	logs = strings.Replace(logs, "ENTERENTER", "\n", -1)
	logs = strings.Replace(logs, "ENTER", "\n", -1)
	logs = strings.Replace(logs, "TABTAB", "<TAB>", -1)
	logs = strings.Replace(logs, "TAB", "<TAB>", -1)
	logs = strings.Replace(logs, "ESCESC", "<ESC>", -1)
	logs = strings.Replace(logs, "ESC", "<ESC>", -1)
	logs = strings.Replace(logs, "CAPS_LOCKCAPS_LOCK", "<CPS_LCK>", -1)
	logs = strings.Replace(logs, "CAPS_LOCK", "<CPS_LCK>", -1)
	logs = strings.Replace(logs, "BSBS", "<BKS>", -1)
	logs = strings.Replace(logs, "BS", "<BKS>", -1)
	logs = strings.Replace(logs, "Del", " <DEL> ", -1)
	logs = strings.Replace(logs, "UpUp", "↑", -1)
	logs = strings.Replace(logs, "Up", "↑", -1)
	logs = strings.Replace(logs, "DownDown", "↓", -1)
	logs = strings.Replace(logs, "Down", "↓", -1)
	logs = strings.Replace(logs, "LeftLeft", "←", -1)
	logs = strings.Replace(logs, "Left", "←", -1)
	logs = strings.Replace(logs, "RightRight", "→", -1)
	logs = strings.Replace(logs, "Right", "→", -1)

	//Alphabets
	logs = strings.Replace(logs, "AA", "A", -1)
	logs = strings.Replace(logs, "BB", "B", -1)
	logs = strings.Replace(logs, "CC", "C", -1)
	logs = strings.Replace(logs, "DD", "D", -1)
	logs = strings.Replace(logs, "EE", "E", -1)
	logs = strings.Replace(logs, "FF", "F", -1)
	logs = strings.Replace(logs, "GG", "G", -1)
	logs = strings.Replace(logs, "HH", "H", -1)
	logs = strings.Replace(logs, "II", "I", -1)
	logs = strings.Replace(logs, "JJ", "J", -1)
	logs = strings.Replace(logs, "KK", "K", -1)
	logs = strings.Replace(logs, "LL", "L", -1)
	logs = strings.Replace(logs, "MM", "M", -1)
	logs = strings.Replace(logs, "NN", "N", -1)
	logs = strings.Replace(logs, "OO", "O", -1)
	logs = strings.Replace(logs, "PP", "P", -1)
	logs = strings.Replace(logs, "QQ", "Q", -1)
	logs = strings.Replace(logs, "RR", "R", -1)
	logs = strings.Replace(logs, "SS", "S", -1)
	logs = strings.Replace(logs, "TT", "T", -1)
	logs = strings.Replace(logs, "UU", "U", -1)
	logs = strings.Replace(logs, "VV", "V", -1)
	logs = strings.Replace(logs, "WW", "W", -1)
	logs = strings.Replace(logs, "XX", "X", -1)
	logs = strings.Replace(logs, "YY", "Y", -1)
	logs = strings.Replace(logs, "ZZ", "Z", -1)

	logs = strings.Replace(logs, "11", "1", -1)
	logs = strings.Replace(logs, "22", "2", -1)
	logs = strings.Replace(logs, "33", "3", -1)
	logs = strings.Replace(logs, "44", "4", -1)
	logs = strings.Replace(logs, "55", "5", -1)
	logs = strings.Replace(logs, "66", "6", -1)
	logs = strings.Replace(logs, "77", "7", -1)
	logs = strings.Replace(logs, "88", "8", -1)
	logs = strings.Replace(logs, "99", "9", -1)
	logs = strings.Replace(logs, "00", "0", -1)

	logs = strings.Replace(logs, "!!", "!", -1)
	logs = strings.Replace(logs, "@@", "@", -1)
	logs = strings.Replace(logs, "##", "#", -1)
	logs = strings.Replace(logs, "$$", "$", -1)
	logs = strings.Replace(logs, "%%", "%", -1)
	logs = strings.Replace(logs, "^^", "^", -1)
	logs = strings.Replace(logs, "&&", "&", -1)
	logs = strings.Replace(logs, "**", "*", -1)
	logs = strings.Replace(logs, "((", "(", -1)
	logs = strings.Replace(logs, "))", ")", -1)
	logs = strings.Replace(logs, "--", "-", -1)
	logs = strings.Replace(logs, "__", "_", -1)
	logs = strings.Replace(logs, "==", "=", -1)
	logs = strings.Replace(logs, "++", "+", -1)
	logs = strings.Replace(logs, "||", "|", -1)
	logs = strings.Replace(logs, "\\\\", "\\", -1)
	logs = strings.Replace(logs, "::", ":", -1)
	logs = strings.Replace(logs, ";;", ";", -1)
	logs = strings.Replace(logs, "\"\"", "\"", -1)
	logs = strings.Replace(logs, "''", "'", -1)
	logs = strings.Replace(logs, "<<", "<", -1)
	logs = strings.Replace(logs, ">>", ">", -1)
	logs = strings.Replace(logs, ",,", ",", -1)
	logs = strings.Replace(logs, "..", ".", -1)
	logs = strings.Replace(logs, "??", "?", -1)
	logs = strings.Replace(logs, "//", "/", -1)
	logs = strings.Replace(logs, "``", "`", -1)
	logs = strings.Replace(logs, "~~", "~", -1)

	return logs
}

/*
efficient
func FormatKeyLogs(logs string) string {
	replacements := map[string]string{
		"Z": "Z",
		"1": "1",
		"2": "2",
	}

	for key, value := range replacements {
		logs = strings.Replace(logs, key, value, -1)
	}

	return logs
}
*/
