package termtex

// Extended symbol tables, merged into symbolCommands / operatorCommands
// / textCommands at init. Kept separate from parser.go so the core
// tables there stay readable; coverage here follows KaTeX's supported
// symbol list. Atom classes for the new glyphs live in spacing.go and
// ASCII fallbacks in ascii.go.

// extraSymbols are Ord atoms: letters, constants, and miscellany.
var extraSymbols = map[string]string{
	// Greek variants (LaTeX \epsilon/\phi are the lunate/closed forms;
	// termtex already uses the common ε/φ for those, so the var- forms
	// map to the same glyphs rather than swapping them).
	"varepsilon": "ε", "vartheta": "ϑ", "varpi": "ϖ", "varrho": "ϱ",
	"varsigma": "ς", "varphi": "φ", "varkappa": "ϰ", "digamma": "ϝ",
	// Uppercase Greek that coincides with Latin capitals.
	"Alpha": "A", "Beta": "B", "Epsilon": "E", "Zeta": "Z", "Eta": "H",
	"Iota": "I", "Kappa": "K", "Mu": "M", "Nu": "N", "Omicron": "O",
	"Rho": "P", "Tau": "T", "Chi": "X",
	// Letterlike
	"emptyset": "∅", "varnothing": "∅", "aleph": "ℵ", "beth": "ℶ", "gimel": "ℷ",
	"daleth": "ℸ", "Re": "ℜ", "Im": "ℑ", "wp": "℘", "imath": "ı", "jmath": "ȷ",
	"hslash": "ℏ", "mho": "℧", "eth": "ð", "Finv": "Ⅎ", "Game": "⅁",
	"complement": "∁", "backslash": "∖", "surd": "√",
	// Geometry and misc
	"angle": "∠", "measuredangle": "∡", "sphericalangle": "∢",
	"triangle": "△", "square": "□", "Box": "□", "blacksquare": "■",
	"Diamond": "◇", "lozenge": "◊", "blacklozenge": "⧫", "bigstar": "★",
	"clubsuit": "♣", "diamondsuit": "♢", "heartsuit": "♡", "spadesuit": "♠",
	"flat": "♭", "natural": "♮", "sharp": "♯",
	"neg": "¬", "lnot": "¬", "top": "⊤", "bot": "⊥",
	"S": "§", "P": "¶", "copyright": "©", "circledR": "®", "pounds": "£",
	"mathsterling": "£", "degree": "°", "textdegree": "°", "checkmark": "✓",
	"maltese": "✠", "yen": "¥", "euro": "€",
	"dprime": "″", "backprime": "‵", "ldotp": ".", "cdotp": "·",
	"infty": "∞",
}

// extraOperators are Bin, Rel, Open/Close and Inner atoms; their class
// comes from the glyph tables in spacing.go.
var extraOperators = map[string]string{
	// Binary
	"oplus": "⊕", "ominus": "⊖", "otimes": "⊗", "oslash": "⊘", "odot": "⊙",
	"circledcirc": "⊚", "circledast": "⊛", "circleddash": "⊝",
	"boxplus": "⊞", "boxminus": "⊟", "boxtimes": "⊠", "boxdot": "⊡",
	"land": "∧", "wedge": "∧", "lor": "∨", "vee": "∨",
	"sqcap": "⊓", "sqcup": "⊔", "uplus": "⊎", "amalg": "⨿", "wr": "≀",
	"bigcirc": "◯", "bigtriangleup": "△", "bigtriangledown": "▽",
	"triangleleft": "◁", "triangleright": "▷", "lhd": "⊲", "rhd": "⊳",
	"unlhd": "⊴", "unrhd": "⊵", "dagger": "†", "ddagger": "‡", "diamond": "⋄",
	"centerdot": "⋅", "ltimes": "⋉", "rtimes": "⋊",
	"leftthreetimes": "⋋", "rightthreetimes": "⋌", "curlyvee": "⋎",
	"curlywedge": "⋏", "veebar": "⊻", "barwedge": "⊼", "doublebarwedge": "⌆",
	"divideontimes": "⋇", "dotplus": "∔", "intercal": "⊺",
	"smallsetminus": "∖", "cdot": "·", "bullet": "•",
	// Relations
	"sim": "∼", "simeq": "≃", "cong": "≅", "ncong": "≇", "nsim": "≁",
	"propto": "∝", "doteq": "≐", "doteqdot": "≑", "triangleq": "≜",
	"approxeq": "≊", "asymp": "≍", "bowtie": "⋈", "eqcirc": "≖", "circeq": "≗",
	"risingdotseq": "≓", "fallingdotseq": "≒", "backsim": "∽", "eqsim": "≂",
	"models": "⊨", "vdash": "⊢", "dashv": "⊣", "Vdash": "⊩", "vDash": "⊨",
	"Vvdash": "⊪", "nvdash": "⊬", "nvDash": "⊭",
	"perp": "⊥", "parallel": "∥", "nparallel": "∦", "nmid": "∤", "shortmid": "∣",
	"ll": "≪", "gg": "≫", "lll": "⋘", "ggg": "⋙", "lt": "<", "gt": ">",
	"prec": "≺", "succ": "≻", "preceq": "⪯", "succeq": "⪰",
	"nprec": "⊀", "nsucc": "⊁", "precsim": "≾", "succsim": "≿",
	"subsetneq": "⊊", "supsetneq": "⊋", "supset": "⊃", "supseteq": "⊇",
	"nsubseteq": "⊈", "nsupseteq": "⊉", "nsubset": "⊄", "nsupset": "⊅",
	"Subset": "⋐", "Supset": "⋑",
	"sqsubset": "⊏", "sqsupset": "⊐", "sqsubseteq": "⊑", "sqsupseteq": "⊒",
	"ni": "∋", "owns": "∋", "notni": "∌",
	"nless": "≮", "ngtr": "≯", "nleq": "≰", "ngeq": "≱", "nle": "≰", "nge": "≱",
	"leqq": "≦", "geqq": "≧", "leqslant": "⩽", "geqslant": "⩾",
	"lesssim": "≲", "gtrsim": "≳", "lessgtr": "≶", "gtrless": "≷",
	"lessapprox": "⪅", "gtrapprox": "⪆", "lesseqgtr": "⋚", "gtreqless": "⋛",
	"between": "≬", "pitchfork": "⋔", "smile": "⌣", "frown": "⌢",
	"vartriangleleft": "⊲", "vartriangleright": "⊳",
	"trianglelefteq": "⊴", "trianglerighteq": "⊵",
	"therefore": "∴", "because": "∵", "coloneqq": "≔", "eqqcolon": "≕",
	"coloneq": "≔", "eqcolon": "≕", "equiv": "≡",
	// Arrows
	"iff": "⟺", "implies": "⟹", "impliedby": "⟸",
	"Longleftrightarrow": "⟺", "Longrightarrow": "⟹", "Longleftarrow": "⟸",
	"longrightarrow": "⟶", "longleftarrow": "⟵", "longleftrightarrow": "⟷",
	"longmapsto": "⟼", "mapsto": "↦", "gets": "←",
	"hookrightarrow": "↪", "hookleftarrow": "↩",
	"twoheadrightarrow": "↠", "twoheadleftarrow": "↞",
	"rightharpoonup": "⇀", "rightharpoondown": "⇁",
	"leftharpoonup": "↼", "leftharpoondown": "↽",
	"rightleftharpoons": "⇌", "leftrightharpoons": "⇋",
	"rightleftarrows": "⇄", "leftrightarrows": "⇆",
	"rightrightarrows": "⇉", "leftleftarrows": "⇇",
	"uparrow": "↑", "downarrow": "↓", "updownarrow": "↕",
	"Uparrow": "⇑", "Downarrow": "⇓", "Updownarrow": "⇕",
	"nearrow": "↗", "searrow": "↘", "swarrow": "↙", "nwarrow": "↖",
	"leadsto": "⇝", "rightsquigarrow": "⇝", "leftrightsquigarrow": "↭",
	"nrightarrow": "↛", "nleftarrow": "↚", "nRightarrow": "⇏", "nLeftarrow": "⇍",
	"nleftrightarrow": "↮", "nLeftrightarrow": "⇎",
	"circlearrowleft": "↺", "circlearrowright": "↻",
	"curvearrowleft": "↶", "curvearrowright": "↷",
	"Lsh": "↰", "Rsh": "↱", "dashrightarrow": "⇢", "dashleftarrow": "⇠",
	// Delimiters
	"lbrace": "{", "rbrace": "}", "lbrack": "[", "rbrack": "]",
	"lgroup": "(", "rgroup": ")", "ulcorner": "⌜", "urcorner": "⌝",
	"llcorner": "⌞", "lrcorner": "⌟",
	// Dots
	"dotsc": "…", "dotso": "…", "dotsb": "⋯", "dotsm": "⋯", "dotsi": "⋯",
	"iddots": "⋰", "adots": "⋰",
}

// extraFunctions are named operators set upright without limits.
var extraFunctions = []string{
	"arg", "coth", "sech", "csch", "arccot", "arcsec", "arccsc",
	"arsinh", "arcosh", "artanh", "lg", "sgn", "tr", "rank", "span",
	"erf", "erfc", "sinc", "Var", "Cov", "lcm", "sign",
}

// limitOps are named operators whose scripts are limits: stacked
// above/below in display style, like \sum.
var limitOps = map[string]bool{
	"lim": true, "limsup": true, "liminf": true, "varlimsup": true,
	"varliminf": true, "injlim": true, "projlim": true,
	"max": true, "min": true, "sup": true, "inf": true,
	"argmax": true, "argmin": true, "det": true, "gcd": true, "Pr": true,
}

// negations maps a relation glyph to its slashed form for \not.
// Anything not listed gets a combining long solidus (U+0338).
var negations = map[string]string{
	"=": "≠", "<": "≮", ">": "≯", "≤": "≰", "≥": "≱", "∈": "∉", "∋": "∌",
	"⊂": "⊄", "⊃": "⊅", "⊆": "⊈", "⊇": "⊉", "≡": "≢", "∼": "≁", "≃": "≄",
	"≅": "≇", "≈": "≉", "|": "∤", "∣": "∤", "∥": "∦", "→": "↛", "←": "↚",
	"⇒": "⇏", "⇐": "⇍", "↔": "↮", "⇔": "⇎", "≺": "⊀", "≻": "⊁", "⊢": "⊬",
	"⊨": "⊭", "∃": "∄",
}

// sizeCommands are \big-style delimiter sizes; termtex sizes all
// delimiters from their content, so these are no-ops.
var sizeCommands = map[string]bool{
	"big": true, "Big": true, "bigg": true, "Bigg": true,
	"bigl": true, "bigr": true, "bigm": true, "Bigl": true, "Bigr": true, "Bigm": true,
	"biggl": true, "biggr": true, "biggm": true, "Biggl": true, "Biggr": true, "Biggm": true,
}

func init() {
	for k, v := range extraSymbols {
		symbolCommands[k] = v
	}
	for k, v := range extraOperators {
		operatorCommands[k] = v
	}
	for _, f := range extraFunctions {
		textCommands[f] = struct{}{}
	}
}
