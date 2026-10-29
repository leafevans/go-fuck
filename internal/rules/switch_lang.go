package rules

import (
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// switchLangTargetLayout is the QWERTY layout the command is converted to.
const switchLangTargetLayout = "qwertyuiop[]asdfghjkl;'zxcvbnm,./QWERTYUIOP{}ASDFGHJKL:\"ZXCVBNM<>?"

// switchLangGreekLayout is the Greek keyboard layout.
const switchLangGreekLayout = ";\u03c2\u03b5\u03c1\u03c4\u03c5\u03b8\u03b9\u03bf\u03c0[]\u03b1\u03c3\u03b4\u03c6\u03b3\u03b7\u03be\u03ba\u03bb\u0384\u03b6\u03c7\u03c8\u03c9\u03b2\u03bd\u03bc,./:\u0385\u0395\u03a1\u03a4\u03a5\u0398\u0399\u039f\u03a0{}\u0391\u03a3\u0394\u03a6\u0393\u0397\u039e\u039a\u039b\u00a8\"\u0396\u03a7\u03a8\u03a9\u0392\u039d\u039c<>?"

// switchLangKoreanLayout is the Korean Dubeolsik keyboard layout.
const switchLangKoreanLayout = "\u3142\u3148\u3137\u3131\u3145\u315b\u3155\u3151\u3150\u3154[]\u3141\u3134\u3147\u3139\u314e\u3157\u3153\u314f\u3163;'\u314b\u314c\u314a\u314d\u3160\u315c\u3161,./\u3143\u3149\u3138\u3132\u3146\u315b\u3155\u3151\u3152\u3156{}\u3141\u3134\u3147\u3139\u314e\u3157\u3153\u314f\u3163:\"\u314b\u314c\u314a\u314d\u3160\u315c\u3161<>?"

// switchLangGreekSourceToTarget maps the Greek layout to the target
// layout. The mapping is explicit because the layout contains accented
// letters that have no position of their own.
var switchLangGreekSourceToTarget = map[rune]rune{
	'\u003b': '\u0071',
	'\u03c2': '\u0077',
	'\u03b5': '\u0065',
	'\u03c1': '\u0072',
	'\u03c4': '\u0074',
	'\u03c5': '\u0079',
	'\u03b8': '\u0075',
	'\u03b9': '\u0069',
	'\u03bf': '\u006f',
	'\u03c0': '\u0070',
	'\u005b': '\u005b',
	'\u005d': '\u005d',
	'\u03b1': '\u0061',
	'\u03c3': '\u0073',
	'\u03b4': '\u0064',
	'\u03c6': '\u0066',
	'\u03b3': '\u0067',
	'\u03b7': '\u0068',
	'\u03be': '\u006a',
	'\u03ba': '\u006b',
	'\u03bb': '\u006c',
	'\u0384': '\u0027',
	'\u03b6': '\u007a',
	'\u03c7': '\u0078',
	'\u03c8': '\u0063',
	'\u03c9': '\u0076',
	'\u03b2': '\u0062',
	'\u03bd': '\u006e',
	'\u03bc': '\u006d',
	'\u002c': '\u002c',
	'\u002e': '\u002e',
	'\u002f': '\u002f',
	'\u003a': '\u0051',
	'\u0385': '\u0057',
	'\u0395': '\u0045',
	'\u03a1': '\u0052',
	'\u03a4': '\u0054',
	'\u03a5': '\u0059',
	'\u0398': '\u0055',
	'\u0399': '\u0049',
	'\u039f': '\u004f',
	'\u03a0': '\u0050',
	'\u007b': '\u007b',
	'\u007d': '\u007d',
	'\u0391': '\u0041',
	'\u03a3': '\u0053',
	'\u0394': '\u0044',
	'\u03a6': '\u0046',
	'\u0393': '\u0047',
	'\u0397': '\u0048',
	'\u039e': '\u004a',
	'\u039a': '\u004b',
	'\u039b': '\u004c',
	'\u00a8': '\u003a',
	'\u0022': '\u0022',
	'\u0396': '\u005a',
	'\u03a7': '\u0058',
	'\u03a8': '\u0043',
	'\u03a9': '\u0056',
	'\u0392': '\u0042',
	'\u039d': '\u004e',
	'\u039c': '\u004d',
	'\u003c': '\u003c',
	'\u003e': '\u003e',
	'\u003f': '\u003f',
	'\u03ac': '\u0061',
	'\u03ad': '\u0065',
	'\u03cd': '\u0079',
	'\u03af': '\u0069',
	'\u03cc': '\u006f',
	'\u03ae': '\u0068',
	'\u03ce': '\u0076',
	'\u0386': '\u0041',
	'\u0388': '\u0045',
	'\u038e': '\u0059',
	'\u038a': '\u0049',
	'\u038c': '\u004f',
	'\u0389': '\u0048',
	'\u038f': '\u0056',
}

// switchLangSourceLayouts lists the recognized keyboard layouts, in the
// order they are tried.
var switchLangSourceLayouts = []switchLangLayout{
	{layout: "\u0439\u0446\u0443\u043a\u0435\u043d\u0433\u0448\u0449\u0437\u0445\u044a\u0444\u044b\u0432\u0430\u043f\u0440\u043e\u043b\u0434\u0436\u044d\u044f\u0447\u0441\u043c\u0438\u0442\u044c\u0431\u044e.\u0419\u0426\u0423\u041a\u0415\u041d\u0413\u0428\u0429\u0417\u0425\u042a\u0424\u042b\u0412\u0410\u041f\u0420\u041e\u041b\u0414\u0416\u042d\u042f\u0427\u0421\u041c\u0418\u0422\u042c\u0411\u042e,"},
	{layout: "\u0439\u0446\u0443\u043a\u0435\u043d\u0433\u0448\u0449\u0437\u0445\u0457\u0444\u0456\u0432\u0430\u043f\u0440\u043e\u043b\u0434\u0436\u0454\u044f\u0447\u0441\u043c\u0438\u0442\u044c\u0431\u044e.\u0419\u0426\u0423\u041a\u0415\u041d\u0413\u0428\u0429\u0417\u0425\u0407\u0424\u0406\u0412\u0410\u041f\u0420\u041e\u041b\u0414\u0416\u0404\u042f\u0427\u0421\u041c\u0418\u0422\u042c\u0411\u042e,"},
	{layout: "\u0636\u0635\u062b\u0642\u0641\u063a\u0639\u0647\u062e\u062d\u062c\u0686\u0634\u0633\u06cc\u0628\u0644\u0627\u062a\u0646\u0645\u06a9\u06af\u0638\u0637\u0632\u0631\u0630\u062f\u067e\u0648./\u0652\u064c\u064d\u064b\u064f\u0650\u064e\u0651][}{\u0624\u0626\u064a\u0625\u0623\u0622\u0629\u00bb\u00ab:\u061b\u0643\u0653\u0698\u0670\u200c\u0654\u0621><\u061f"},
	{layout: "/'\u05e7\u05e8\u05d0\u05d8\u05d5\u05df\u05dd\u05e4][\u05e9\u05d3\u05d2\u05db\u05e2\u05d9\u05d7\u05dc\u05da\u05e3,\u05d6\u05e1\u05d1\u05d4\u05e0\u05de\u05e6\u05ea\u05e5.QWERTYUIOP{}ASDFGHJKL:\"ZXCVBNM<>?"},
	{layout: switchLangGreekLayout, mapping: switchLangGreekSourceToTarget},
	{layout: switchLangKoreanLayout},
}

// switchLangKoreanHeads lists the leading consonants of Hangul syllables.
var switchLangKoreanHeads = []rune{'\u3131', '\u3132', '\u3134', '\u3137', '\u3138', '\u3139', '\u3141', '\u3142', '\u3143', '\u3145', '\u3146', '\u3147', '\u3148', '\u3149', '\u314a', '\u314b', '\u314c', '\u314d', '\u314e'}

// switchLangKoreanBodies lists the vowels of Hangul syllables.
var switchLangKoreanBodies = []rune{'\u314f', '\u3150', '\u3151', '\u3152', '\u3153', '\u3154', '\u3155', '\u3156', '\u3157', '\u3158', '\u3159', '\u315a', '\u315b', '\u315c', '\u315d', '\u315e', '\u315f', '\u3160', '\u3161', '\u3162', '\u3163'}

// switchLangKoreanTails lists the trailing consonants of Hangul syllables.
var switchLangKoreanTails = []rune{'\u0020', '\u3131', '\u3132', '\u3133', '\u3134', '\u3135', '\u3136', '\u3137', '\u3139', '\u313a', '\u313b', '\u313c', '\u313d', '\u313e', '\u313f', '\u3140', '\u3141', '\u3142', '\u3144', '\u3145', '\u3146', '\u3147', '\u3148', '\u314a', '\u314b', '\u314c', '\u314d', '\u314e'}

// switchLangKoreanDoubles lists the jamo that are written with two
// letters.
var switchLangKoreanDoubles = []string{"\u3158", "\u3159", "\u315a", "\u315d", "\u315e", "\u315f", "\u3162", "\u3133", "\u3135", "\u3136", "\u313a", "\u313b", "\u313c", "\u313d", "\u313e", "\u3140", "\u3144"}

// switchLangKoreanDoubleMods lists the two letter forms of the double
// jamo.
var switchLangKoreanDoubleMods = []string{"\u3157\u314f", "\u3157\u3150", "\u3157\u3163", "\u315c\u3153", "\u315c\u3154", "\u315c\u3163", "\u3161\u3163", "\u3131\u3145", "\u3134\u3148", "\u3134\u314e", "\u3139\u3131", "\u3139\u3141", "\u3139\u3142", "\u3139\u3145", "\u3139\u314c", "\u3139\u314e", "\u3142\u3145"}

// switchLangTargetRunes is the target layout as a rune slice.
var switchLangTargetRunes = []rune(switchLangTargetLayout)

// switchLangSourceRunes holds the source layouts as rune slices.
var switchLangSourceRunes = func() [][]rune {
	runes := make([][]rune, len(switchLangSourceLayouts))
	for index, source := range switchLangSourceLayouts {
		runes[index] = []rune(source.layout)
	}
	return runes
}()

// switchLangLayout describes one source keyboard layout.
type switchLangLayout struct {
	// layout is the layout itself.
	layout string
	// mapping explicitly maps characters; when nil the layout is converted by
	// position.
	mapping map[rune]rune
}

// switchLangContains reports whether the layout contains the character.
func switchLangContains(layout []rune, ch rune) bool {
	return slices.Contains(layout, ch)
}

// switchLangMatchesLayout reports whether every character of the command
// belongs to the layout. Dashes and underscores are ignored because they exist
// on every keyboard.
func switchLangMatchesLayout(script string, layout []rune) bool {
	for _, part := range strings.Split(script, " ") {
		for _, ch := range part {
			if ch == '-' || ch == '_' {
				continue
			}
			if !switchLangContains(layout, ch) {
				return false
			}
		}
	}
	return true
}

// switchLangMatchedLayout returns the layout the command was typed in, or an
// empty string when no layout matches.
func switchLangMatchedLayout(script string) string {
	for index, source := range switchLangSourceLayouts {
		if switchLangMatchesLayout(script, switchLangSourceRunes[index]) {
			return source.layout
		}
	}
	return ""
}

// switchLangLayoutIndex returns the index of the layout, or -1.
func switchLangLayoutIndex(layout string) int {
	for index, source := range switchLangSourceLayouts {
		if source.layout == layout {
			return index
		}
	}
	return -1
}

// switchLangSwitch converts a single character by its position in the layout.
func switchLangSwitch(ch rune, layout []rune) rune {
	for index, candidate := range layout {
		if candidate == ch {
			return switchLangTargetRunes[index]
		}
	}
	return ch
}

// switchLangSwitchCommand converts the command from the layout to the target
// layout.
func switchLangSwitchCommand(script, layout string) string {
	index := switchLangLayoutIndex(layout)
	if index < 0 {
		return script
	}
	source := switchLangSourceLayouts[index]
	layoutRunes := switchLangSourceRunes[index]

	converted := make([]rune, 0, len(script))
	for _, ch := range script {
		if source.mapping != nil {
			if target, ok := source.mapping[ch]; ok {
				converted = append(converted, target)
			} else {
				converted = append(converted, ch)
			}
			continue
		}
		converted = append(converted, switchLangSwitch(ch, layoutRunes))
	}
	return string(converted)
}

// Hangul syllable constants. A syllable is composed of a leading consonant, a
// vowel and an optional trailing consonant.
const (
	switchLangHangulFirst     = 0xAC00
	switchLangHangulLast      = 0xD7A3
	switchLangHangulHeadCount = 588
	switchLangHangulBodyCount = 28
)

// switchLangIsKorean reports whether the character is a Korean letter.
func switchLangIsKorean(ch rune) bool {
	return (ch >= 0x3131 && ch <= 0x314E) || // jamo consonants and vowels
		(ch >= 0x314F && ch <= 0x3163) ||
		(ch >= switchLangHangulFirst && ch <= switchLangHangulLast)
}

// switchLangChangeDouble replaces a double jamo with its two letter form.
func switchLangChangeDouble(ch rune) string {
	for index, candidate := range switchLangKoreanDoubles {
		if candidate == string(ch) {
			return switchLangKoreanDoubleMods[index]
		}
	}
	return string(ch)
}

// switchLangDecomposeKorean splits Hangul syllables into individual letters.
func switchLangDecomposeKorean(script string) string {
	var builder strings.Builder
	for _, ch := range script {
		if ch < switchLangHangulFirst || ch > switchLangHangulLast {
			builder.WriteString(switchLangChangeDouble(ch))
			continue
		}

		offset := int(ch) - switchLangHangulFirst
		head := offset / switchLangHangulHeadCount
		body := (offset - switchLangHangulHeadCount*head) / switchLangHangulBodyCount
		tail := offset - switchLangHangulHeadCount*head - switchLangHangulBodyCount*body

		for _, letter := range []rune{
			switchLangKoreanHeads[head],
			switchLangKoreanBodies[body],
			switchLangKoreanTails[tail],
		} {
			if letter == ' ' {
				continue
			}
			builder.WriteString(switchLangChangeDouble(letter))
		}
	}
	return builder.String()
}

// switchLangContainsKorean reports whether the command contains Korean letters.
func switchLangContainsKorean(script string) bool {
	for _, ch := range script {
		if switchLangIsKorean(ch) {
			return true
		}
	}
	return false
}

// switchLangMatch reports whether the command was typed with the wrong keyboard
// layout.
func switchLangMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, "not found") {
		return false
	}
	if switchLangContainsKorean(command.Script) {
		return true
	}

	layout := switchLangMatchedLayout(command.Script)
	return layout != "" && switchLangSwitchCommand(command.Script, layout) != utils.GetAlias()
}

// switchLangNewCommand converts the command to the target layout.
func switchLangNewCommand(command *types.Command) ([]string, error) {
	script := command.Script
	if switchLangContainsKorean(script) {
		script = switchLangDecomposeKorean(script)
	}

	layout := switchLangMatchedLayout(script)
	if layout == "" {
		return nil, nil
	}
	return []string{switchLangSwitchCommand(script, layout)}, nil
}

func init() {
	register(types.NewRule("switch_lang", switchLangMatch, switchLangNewCommand))
}
