import { Qn as Subject, Tt as debounceTime, gt as distinctUntilChanged, m as tap, sn as fromEvent, ur as Subscription } from "./esm5-Dur3xngx.js";
import { Cr as inject, I as EventEmitter, Pn as forwardRef, _a as ɵɵresetView, la as ɵɵdefineInjector, va as ɵɵrestoreView } from "./_resource-chunk-BCcMEh28.js";
import { $ as isDevMode, Ai as setClassMetadata, Bo as ɵɵnextContext, Ds as ɵɵstyleMap, Fs as ɵɵtextInterpolate1, Ga as ɵɵdefineNgModule, Gt as ChangeDetectionStrategy, Ji as ɵɵInheritDefinitionFeature, Jo as ɵɵprojectionDef, Jt as Component, Nn as NgModule, Ns as ɵɵtext, Oo as ɵɵinterpolate1, Os as ɵɵstyleProp, Qo as ɵɵpureFunction2, Ro as ɵɵlistener, Sn as Input, Ua as ɵɵdefineComponent, Wa as ɵɵdefineDirective, Xi as ɵɵProvidersFeature, Yi as ɵɵNgOnChangesFeature, Yo as ɵɵproperty, Zi as ɵɵadvance, _s as ɵɵresolveWindow, aa as ɵɵattribute, ca as ɵɵclassMap, co as ɵɵelementEnd, da as ɵɵconditional, ds as ɵɵrepeaterCreate, fn as ElementRef, fs as ɵɵrepeaterTrackByIdentity, go as ɵɵgetCurrentView, hn as HostListener, io as ɵɵelement, lo as ɵɵelementStart, pa as ɵɵconditionalCreate, qo as ɵɵprojection, un as Directive, us as ɵɵrepeater, zn as Output } from "./core-DtssIGqE.js";
import { G as NgStyle, x as CommonModule } from "./common-DgQvbbqr.js";
import { r as NG_VALUE_ACCESSOR } from "./forms-B3gLtxf0.js";
//#region node_modules/.pnpm/@ctrl+tinycolor@4.2.1/node_modules/@ctrl/tinycolor/dist/module/util.js
/**
* Take input from [0, n] and return it as [0, 1]
* @hidden
*/
function bound01(n, max) {
	if (isOnePointZero(n)) n = "100%";
	const isPercent = isPercentage(n);
	n = max === 360 ? n : Math.min(max, Math.max(0, parseFloat(n)));
	if (isPercent) n = parseInt(String(n * max), 10) / 100;
	if (Math.abs(n - max) < 1e-6) return 1;
	if (max === 360) n = (n < 0 ? n % max + max : n % max) / max;
	else n = n % max / max;
	return n;
}
/**
* Force a number between 0 and 1
* @hidden
*/
function clamp01(val) {
	return Math.min(1, Math.max(0, val));
}
/**
* Need to handle 1.0 as 100%, since once it is a number, there is no difference between it and 1
* <http://stackoverflow.com/questions/7422072/javascript-how-to-detect-number-as-a-decimal-including-1-0>
* @hidden
*/
function isOnePointZero(n) {
	return typeof n === "string" && n.indexOf(".") !== -1 && parseFloat(n) === 1;
}
/**
* Check to see if string passed in is a percentage
* @hidden
*/
function isPercentage(n) {
	return typeof n === "string" && n.indexOf("%") !== -1;
}
/**
* Return a valid alpha value [0,1] with all invalid values being set to 1
* @hidden
*/
function boundAlpha(a) {
	a = parseFloat(a);
	if (isNaN(a) || a < 0 || a > 1) a = 1;
	return a;
}
/**
* Replace a decimal with it's percentage value
* @hidden
*/
function convertToPercentage(n) {
	if (Number(n) <= 1) return `${Number(n) * 100}%`;
	return n;
}
/**
* Force a hex value to have 2 characters
* @hidden
*/
function pad2(c) {
	return c.length === 1 ? "0" + c : String(c);
}
//#endregion
//#region node_modules/.pnpm/@ctrl+tinycolor@4.2.1/node_modules/@ctrl/tinycolor/dist/module/conversion.js
/**
* Handle bounds / percentage checking to conform to CSS color spec
* <http://www.w3.org/TR/css3-color/>
* *Assumes:* r, g, b in [0, 255] or [0, 1]
* *Returns:* { r, g, b } in [0, 255]
*/
function rgbToRgb(r, g, b) {
	return {
		r: bound01(r, 255) * 255,
		g: bound01(g, 255) * 255,
		b: bound01(b, 255) * 255
	};
}
/**
* Converts an RGB color value to HSL.
* *Assumes:* r, g, and b are contained in [0, 255] or [0, 1]
* *Returns:* { h, s, l } in [0,1]
*/
function rgbToHsl(r, g, b) {
	r = bound01(r, 255);
	g = bound01(g, 255);
	b = bound01(b, 255);
	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	let h = 0;
	let s = 0;
	const l = (max + min) / 2;
	if (max === min) {
		s = 0;
		h = 0;
	} else {
		const d = max - min;
		s = l > .5 ? d / (2 - max - min) : d / (max + min);
		switch (max) {
			case r:
				h = (g - b) / d + (g < b ? 6 : 0);
				break;
			case g:
				h = (b - r) / d + 2;
				break;
			case b: h = (r - g) / d + 4;
		}
		h /= 6;
	}
	return {
		h,
		s,
		l
	};
}
function hue2rgb(p, q, t) {
	if (t < 0) t += 1;
	if (t > 1) t -= 1;
	if (t < 1 / 6) return p + (q - p) * (6 * t);
	if (t < 1 / 2) return q;
	if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
	return p;
}
/**
* Converts an HSL color value to RGB.
*
* *Assumes:* h is contained in [0, 1] or [0, 360] and s and l are contained [0, 1] or [0, 100]
* *Returns:* { r, g, b } in the set [0, 255]
*/
function hslToRgb(h, s, l) {
	let r;
	let g;
	let b;
	h = bound01(h, 360);
	s = bound01(s, 100);
	l = bound01(l, 100);
	if (s === 0) {
		g = l;
		b = l;
		r = l;
	} else {
		const q = l < .5 ? l * (1 + s) : l + s - l * s;
		const p = 2 * l - q;
		r = hue2rgb(p, q, h + 1 / 3);
		g = hue2rgb(p, q, h);
		b = hue2rgb(p, q, h - 1 / 3);
	}
	return {
		r: r * 255,
		g: g * 255,
		b: b * 255
	};
}
/**
* Converts an RGB color value to HSV
*
* *Assumes:* r, g, and b are contained in the set [0, 255] or [0, 1]
* *Returns:* { h, s, v } in [0,1]
*/
function rgbToHsv(r, g, b) {
	r = bound01(r, 255);
	g = bound01(g, 255);
	b = bound01(b, 255);
	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	let h = 0;
	const v = max;
	const d = max - min;
	const s = max === 0 ? 0 : d / max;
	if (max === min) h = 0;
	else {
		switch (max) {
			case r:
				h = (g - b) / d + (g < b ? 6 : 0);
				break;
			case g:
				h = (b - r) / d + 2;
				break;
			case b: h = (r - g) / d + 4;
		}
		h /= 6;
	}
	return {
		h,
		s,
		v
	};
}
/**
* Converts an HSV color value to RGB.
*
* *Assumes:* h is contained in [0, 1] or [0, 360] and s and v are contained in [0, 1] or [0, 100]
* *Returns:* { r, g, b } in the set [0, 255]
*/
function hsvToRgb(h, s, v) {
	h = bound01(h, 360) * 6;
	s = bound01(s, 100);
	v = bound01(v, 100);
	const i = Math.floor(h);
	const f = h - i;
	const p = v * (1 - s);
	const q = v * (1 - f * s);
	const t = v * (1 - (1 - f) * s);
	const mod = i % 6;
	const r = [
		v,
		q,
		p,
		p,
		t,
		v
	][mod];
	const g = [
		t,
		v,
		v,
		q,
		p,
		p
	][mod];
	const b = [
		p,
		p,
		t,
		v,
		v,
		q
	][mod];
	return {
		r: r * 255,
		g: g * 255,
		b: b * 255
	};
}
/**
* Converts an RGB color to hex
*
* *Assumes:* r, g, and b are contained in the set [0, 255]
* *Returns:* a 3 or 6 character hex
*/
function rgbToHex(r, g, b, allow3Char) {
	const rHex = pad2(Math.round(r).toString(16));
	const gHex = pad2(Math.round(g).toString(16));
	const bHex = pad2(Math.round(b).toString(16));
	if (allow3Char && rHex.startsWith(rHex.charAt(1)) && gHex.startsWith(gHex.charAt(1)) && bHex.startsWith(bHex.charAt(1))) return rHex.charAt(0) + gHex.charAt(0) + bHex.charAt(0);
	return rHex + gHex + bHex;
}
/**
* Converts an RGBA color plus alpha transparency to hex
*
* *Assumes:* r, g, b are contained in the set [0, 255] and a in [0, 1]
* *Returns:* a 4 or 8 character rgba hex
*/
function rgbaToHex(r, g, b, a, allow4Char) {
	const rHex = pad2(Math.round(r).toString(16));
	const gHex = pad2(Math.round(g).toString(16));
	const bHex = pad2(Math.round(b).toString(16));
	const aHex = pad2(convertDecimalToHex(a));
	if (allow4Char && rHex.startsWith(rHex.charAt(1)) && gHex.startsWith(gHex.charAt(1)) && bHex.startsWith(bHex.charAt(1)) && aHex.startsWith(aHex.charAt(1))) return rHex.charAt(0) + gHex.charAt(0) + bHex.charAt(0) + aHex.charAt(0);
	return rHex + gHex + bHex + aHex;
}
/**
* Converts CMYK to RBG
* Assumes c, m, y, k are in the set [0, 100]
*/
function cmykToRgb(c, m, y, k) {
	const cConv = c / 100;
	const mConv = m / 100;
	const yConv = y / 100;
	const kConv = k / 100;
	return {
		r: 255 * (1 - cConv) * (1 - kConv),
		g: 255 * (1 - mConv) * (1 - kConv),
		b: 255 * (1 - yConv) * (1 - kConv)
	};
}
function rgbToCmyk(r, g, b) {
	let c = 1 - r / 255;
	let m = 1 - g / 255;
	let y = 1 - b / 255;
	let k = Math.min(c, m, y);
	if (k === 1) {
		c = 0;
		m = 0;
		y = 0;
	} else {
		c = (c - k) / (1 - k) * 100;
		m = (m - k) / (1 - k) * 100;
		y = (y - k) / (1 - k) * 100;
	}
	k *= 100;
	return {
		c: Math.round(c),
		m: Math.round(m),
		y: Math.round(y),
		k: Math.round(k)
	};
}
/** Converts a decimal to a hex value */
function convertDecimalToHex(d) {
	return Math.round(parseFloat(d) * 255).toString(16);
}
/** Converts a hex value to a decimal */
function convertHexToDecimal(h) {
	return parseIntFromHex(h) / 255;
}
/** Parse a base-16 hex value into a base-10 integer */
function parseIntFromHex(val) {
	return parseInt(val, 16);
}
function numberInputToObject(color) {
	return {
		r: color >> 16,
		g: (color & 65280) >> 8,
		b: color & 255
	};
}
//#endregion
//#region node_modules/.pnpm/@ctrl+tinycolor@4.2.1/node_modules/@ctrl/tinycolor/dist/module/css-color-names.js
/**
* @hidden
*/
var names = {
	aliceblue: "#f0f8ff",
	antiquewhite: "#faebd7",
	aqua: "#00ffff",
	aquamarine: "#7fffd4",
	azure: "#f0ffff",
	beige: "#f5f5dc",
	bisque: "#ffe4c4",
	black: "#000000",
	blanchedalmond: "#ffebcd",
	blue: "#0000ff",
	blueviolet: "#8a2be2",
	brown: "#a52a2a",
	burlywood: "#deb887",
	cadetblue: "#5f9ea0",
	chartreuse: "#7fff00",
	chocolate: "#d2691e",
	coral: "#ff7f50",
	cornflowerblue: "#6495ed",
	cornsilk: "#fff8dc",
	crimson: "#dc143c",
	cyan: "#00ffff",
	darkblue: "#00008b",
	darkcyan: "#008b8b",
	darkgoldenrod: "#b8860b",
	darkgray: "#a9a9a9",
	darkgreen: "#006400",
	darkgrey: "#a9a9a9",
	darkkhaki: "#bdb76b",
	darkmagenta: "#8b008b",
	darkolivegreen: "#556b2f",
	darkorange: "#ff8c00",
	darkorchid: "#9932cc",
	darkred: "#8b0000",
	darksalmon: "#e9967a",
	darkseagreen: "#8fbc8f",
	darkslateblue: "#483d8b",
	darkslategray: "#2f4f4f",
	darkslategrey: "#2f4f4f",
	darkturquoise: "#00ced1",
	darkviolet: "#9400d3",
	deeppink: "#ff1493",
	deepskyblue: "#00bfff",
	dimgray: "#696969",
	dimgrey: "#696969",
	dodgerblue: "#1e90ff",
	firebrick: "#b22222",
	floralwhite: "#fffaf0",
	forestgreen: "#228b22",
	fuchsia: "#ff00ff",
	gainsboro: "#dcdcdc",
	ghostwhite: "#f8f8ff",
	goldenrod: "#daa520",
	gold: "#ffd700",
	gray: "#808080",
	green: "#008000",
	greenyellow: "#adff2f",
	grey: "#808080",
	honeydew: "#f0fff0",
	hotpink: "#ff69b4",
	indianred: "#cd5c5c",
	indigo: "#4b0082",
	ivory: "#fffff0",
	khaki: "#f0e68c",
	lavenderblush: "#fff0f5",
	lavender: "#e6e6fa",
	lawngreen: "#7cfc00",
	lemonchiffon: "#fffacd",
	lightblue: "#add8e6",
	lightcoral: "#f08080",
	lightcyan: "#e0ffff",
	lightgoldenrodyellow: "#fafad2",
	lightgray: "#d3d3d3",
	lightgreen: "#90ee90",
	lightgrey: "#d3d3d3",
	lightpink: "#ffb6c1",
	lightsalmon: "#ffa07a",
	lightseagreen: "#20b2aa",
	lightskyblue: "#87cefa",
	lightslategray: "#778899",
	lightslategrey: "#778899",
	lightsteelblue: "#b0c4de",
	lightyellow: "#ffffe0",
	lime: "#00ff00",
	limegreen: "#32cd32",
	linen: "#faf0e6",
	magenta: "#ff00ff",
	maroon: "#800000",
	mediumaquamarine: "#66cdaa",
	mediumblue: "#0000cd",
	mediumorchid: "#ba55d3",
	mediumpurple: "#9370db",
	mediumseagreen: "#3cb371",
	mediumslateblue: "#7b68ee",
	mediumspringgreen: "#00fa9a",
	mediumturquoise: "#48d1cc",
	mediumvioletred: "#c71585",
	midnightblue: "#191970",
	mintcream: "#f5fffa",
	mistyrose: "#ffe4e1",
	moccasin: "#ffe4b5",
	navajowhite: "#ffdead",
	navy: "#000080",
	oldlace: "#fdf5e6",
	olive: "#808000",
	olivedrab: "#6b8e23",
	orange: "#ffa500",
	orangered: "#ff4500",
	orchid: "#da70d6",
	palegoldenrod: "#eee8aa",
	palegreen: "#98fb98",
	paleturquoise: "#afeeee",
	palevioletred: "#db7093",
	papayawhip: "#ffefd5",
	peachpuff: "#ffdab9",
	peru: "#cd853f",
	pink: "#ffc0cb",
	plum: "#dda0dd",
	powderblue: "#b0e0e6",
	purple: "#800080",
	rebeccapurple: "#663399",
	red: "#ff0000",
	rosybrown: "#bc8f8f",
	royalblue: "#4169e1",
	saddlebrown: "#8b4513",
	salmon: "#fa8072",
	sandybrown: "#f4a460",
	seagreen: "#2e8b57",
	seashell: "#fff5ee",
	sienna: "#a0522d",
	silver: "#c0c0c0",
	skyblue: "#87ceeb",
	slateblue: "#6a5acd",
	slategray: "#708090",
	slategrey: "#708090",
	snow: "#fffafa",
	springgreen: "#00ff7f",
	steelblue: "#4682b4",
	tan: "#d2b48c",
	teal: "#008080",
	thistle: "#d8bfd8",
	tomato: "#ff6347",
	turquoise: "#40e0d0",
	violet: "#ee82ee",
	wheat: "#f5deb3",
	white: "#ffffff",
	whitesmoke: "#f5f5f5",
	yellow: "#ffff00",
	yellowgreen: "#9acd32"
};
//#endregion
//#region node_modules/.pnpm/@ctrl+tinycolor@4.2.1/node_modules/@ctrl/tinycolor/dist/module/format-input.js
/**
* Given a string or object, convert that input to RGB
*
* Possible string inputs:
* ```
* "red"
* "#f00" or "f00"
* "#ff0000" or "ff0000"
* "#ff000000" or "ff000000"
* "rgb 255 0 0" or "rgb (255, 0, 0)"
* "rgb 1.0 0 0" or "rgb (1, 0, 0)"
* "rgba (255, 0, 0, 1)" or "rgba 255, 0, 0, 1"
* "rgba (1.0, 0, 0, 1)" or "rgba 1.0, 0, 0, 1"
* "hsl(0, 100%, 50%)" or "hsl 0 100% 50%"
* "hsla(0, 100%, 50%, 1)" or "hsla 0 100% 50%, 1"
* "hsv(0, 100%, 100%)" or "hsv 0 100% 100%"
* "cmyk(0, 20, 0, 0)" or "cmyk 0 20 0 0"
* ```
*/
function inputToRGB(color) {
	let rgb = {
		r: 0,
		g: 0,
		b: 0
	};
	let a = 1;
	let s = null;
	let v = null;
	let l = null;
	let ok = false;
	let format = false;
	if (typeof color === "string") color = stringInputToObject(color);
	if (typeof color === "object") {
		if (isValidCSSUnit(color.r) && isValidCSSUnit(color.g) && isValidCSSUnit(color.b)) {
			rgb = rgbToRgb(color.r, color.g, color.b);
			ok = true;
			format = String(color.r).substr(-1) === "%" ? "prgb" : "rgb";
		} else if (isValidCSSUnit(color.h) && isValidCSSUnit(color.s) && isValidCSSUnit(color.v)) {
			s = convertToPercentage(color.s);
			v = convertToPercentage(color.v);
			rgb = hsvToRgb(color.h, s, v);
			ok = true;
			format = "hsv";
		} else if (isValidCSSUnit(color.h) && isValidCSSUnit(color.s) && isValidCSSUnit(color.l)) {
			s = convertToPercentage(color.s);
			l = convertToPercentage(color.l);
			rgb = hslToRgb(color.h, s, l);
			ok = true;
			format = "hsl";
		} else if (isValidCSSUnit(color.c) && isValidCSSUnit(color.m) && isValidCSSUnit(color.y) && isValidCSSUnit(color.k)) {
			rgb = cmykToRgb(color.c, color.m, color.y, color.k);
			ok = true;
			format = "cmyk";
		}
		if (Object.prototype.hasOwnProperty.call(color, "a")) a = color.a;
	}
	a = boundAlpha(a);
	return {
		ok,
		format: color.format || format,
		r: Math.min(255, Math.max(rgb.r, 0)),
		g: Math.min(255, Math.max(rgb.g, 0)),
		b: Math.min(255, Math.max(rgb.b, 0)),
		a
	};
}
var CSS_UNIT = "(?:[-\\+]?\\d*\\.\\d+%?)|(?:[-\\+]?\\d+%?)";
var PERMISSIVE_MATCH3 = "[\\s|\\(]+(" + CSS_UNIT + ")[,|\\s]+(" + CSS_UNIT + ")[,|\\s]+(" + CSS_UNIT + ")\\s*\\)?";
var PERMISSIVE_MATCH4 = "[\\s|\\(]+(" + CSS_UNIT + ")[,|\\s]+(" + CSS_UNIT + ")[,|\\s]+(" + CSS_UNIT + ")[,|\\s]+(" + CSS_UNIT + ")\\s*\\)?";
var matchers = {
	hex: /^[0-9a-fA-F]+$/,
	CSS_UNIT: new RegExp(CSS_UNIT),
	rgb: new RegExp("rgb" + PERMISSIVE_MATCH3),
	rgba: new RegExp("rgba" + PERMISSIVE_MATCH4),
	hsl: new RegExp("hsl" + PERMISSIVE_MATCH3),
	hsla: new RegExp("hsla" + PERMISSIVE_MATCH4),
	hsv: new RegExp("hsv" + PERMISSIVE_MATCH3),
	hsva: new RegExp("hsva" + PERMISSIVE_MATCH4),
	cmyk: new RegExp("cmyk" + PERMISSIVE_MATCH4),
	hex3: /^#?([0-9a-fA-F]{1})([0-9a-fA-F]{1})([0-9a-fA-F]{1})$/,
	hex6: /^#?([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$/,
	hex4: /^#?([0-9a-fA-F]{1})([0-9a-fA-F]{1})([0-9a-fA-F]{1})([0-9a-fA-F]{1})$/,
	hex8: /^#?([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$/
};
/**
* Permissive string parsing.  Take in a number of formats, and output an object
* based on detected format.  Returns `{ r, g, b }` or `{ h, s, l }` or `{ h, s, v}` or `{c, m, y, k}` or `{c, m, y, k, a}`
*/
function stringInputToObject(color) {
	color = color.trim().toLowerCase();
	if (color.length === 0) return false;
	let named = false;
	if (names[color]) {
		color = names[color];
		named = true;
	} else if (color === "transparent") return {
		r: 0,
		g: 0,
		b: 0,
		a: 0,
		format: "name"
	};
	let match;
	if (typeof color === "string" && color.length <= 9 && (color.startsWith("#") || matchers.hex.test(color))) {
		match = matchers.hex8.exec(color);
		if (match) return {
			r: parseIntFromHex(match[1]),
			g: parseIntFromHex(match[2]),
			b: parseIntFromHex(match[3]),
			a: convertHexToDecimal(match[4]),
			format: named ? "name" : "hex8"
		};
		match = matchers.hex6.exec(color);
		if (match) return {
			r: parseIntFromHex(match[1]),
			g: parseIntFromHex(match[2]),
			b: parseIntFromHex(match[3]),
			format: named ? "name" : "hex"
		};
		match = matchers.hex4.exec(color);
		if (match) return {
			r: parseIntFromHex(match[1] + match[1]),
			g: parseIntFromHex(match[2] + match[2]),
			b: parseIntFromHex(match[3] + match[3]),
			a: convertHexToDecimal(match[4] + match[4]),
			format: named ? "name" : "hex8"
		};
		match = matchers.hex3.exec(color);
		if (match) return {
			r: parseIntFromHex(match[1] + match[1]),
			g: parseIntFromHex(match[2] + match[2]),
			b: parseIntFromHex(match[3] + match[3]),
			format: named ? "name" : "hex"
		};
	}
	match = matchers.rgb.exec(color);
	if (match) return {
		r: match[1],
		g: match[2],
		b: match[3]
	};
	match = matchers.rgba.exec(color);
	if (match) return {
		r: match[1],
		g: match[2],
		b: match[3],
		a: match[4]
	};
	match = matchers.hsl.exec(color);
	if (match) return {
		h: match[1],
		s: match[2],
		l: match[3]
	};
	match = matchers.hsla.exec(color);
	if (match) return {
		h: match[1],
		s: match[2],
		l: match[3],
		a: match[4]
	};
	match = matchers.hsv.exec(color);
	if (match) return {
		h: match[1],
		s: match[2],
		v: match[3]
	};
	match = matchers.hsva.exec(color);
	if (match) return {
		h: match[1],
		s: match[2],
		v: match[3],
		a: match[4]
	};
	match = matchers.cmyk.exec(color);
	if (match) return {
		c: match[1],
		m: match[2],
		y: match[3],
		k: match[4]
	};
	return false;
}
/**
* Check to see if it looks like a CSS unit
* (see `matchers` above for definition).
*/
function isValidCSSUnit(color) {
	if (typeof color === "number") return !Number.isNaN(color);
	return matchers.CSS_UNIT.test(color);
}
//#endregion
//#region node_modules/.pnpm/@ctrl+tinycolor@4.2.1/node_modules/@ctrl/tinycolor/dist/module/index.js
var TinyColor = class TinyColor {
	constructor(color = "", opts = {}) {
		if (color instanceof TinyColor) return color;
		if (typeof color === "number") color = numberInputToObject(color);
		this.originalInput = color;
		const rgb = inputToRGB(color);
		this.originalInput = color;
		this.r = rgb.r;
		this.g = rgb.g;
		this.b = rgb.b;
		this.a = rgb.a;
		this.roundA = Math.round(100 * this.a) / 100;
		this.format = opts.format ?? rgb.format;
		this.gradientType = opts.gradientType;
		if (this.r < 1) this.r = Math.round(this.r);
		if (this.g < 1) this.g = Math.round(this.g);
		if (this.b < 1) this.b = Math.round(this.b);
		this.isValid = rgb.ok;
	}
	isDark() {
		return this.getBrightness() < 128;
	}
	isLight() {
		return !this.isDark();
	}
	/**
	* Returns the perceived brightness of the color, from 0-255.
	*/
	getBrightness() {
		const rgb = this.toRgb();
		return (rgb.r * 299 + rgb.g * 587 + rgb.b * 114) / 1e3;
	}
	/**
	* Returns the perceived luminance of a color, from 0-1.
	*/
	getLuminance() {
		const rgb = this.toRgb();
		let R;
		let G;
		let B;
		const RsRGB = rgb.r / 255;
		const GsRGB = rgb.g / 255;
		const BsRGB = rgb.b / 255;
		if (RsRGB <= .03928) R = RsRGB / 12.92;
		else R = Math.pow((RsRGB + .055) / 1.055, 2.4);
		if (GsRGB <= .03928) G = GsRGB / 12.92;
		else G = Math.pow((GsRGB + .055) / 1.055, 2.4);
		if (BsRGB <= .03928) B = BsRGB / 12.92;
		else B = Math.pow((BsRGB + .055) / 1.055, 2.4);
		return .2126 * R + .7152 * G + .0722 * B;
	}
	/**
	* Returns the alpha value of a color, from 0-1.
	*/
	getAlpha() {
		return this.a;
	}
	/**
	* Sets the alpha value on the current color.
	*
	* @param alpha - The new alpha value. The accepted range is 0-1.
	*/
	setAlpha(alpha) {
		this.a = boundAlpha(alpha);
		this.roundA = Math.round(100 * this.a) / 100;
		return this;
	}
	/**
	* Returns whether the color is monochrome.
	*/
	isMonochrome() {
		const { s } = this.toHsl();
		return s === 0;
	}
	/**
	* Returns the object as a HSVA object.
	*/
	toHsv() {
		const hsv = rgbToHsv(this.r, this.g, this.b);
		return {
			h: hsv.h * 360,
			s: hsv.s,
			v: hsv.v,
			a: this.a
		};
	}
	/**
	* Returns the hsva values interpolated into a string with the following format:
	* "hsva(xxx, xxx, xxx, xx)".
	*/
	toHsvString() {
		const hsv = rgbToHsv(this.r, this.g, this.b);
		const h = Math.round(hsv.h * 360);
		const s = Math.round(hsv.s * 100);
		const v = Math.round(hsv.v * 100);
		return this.a === 1 ? `hsv(${h}, ${s}%, ${v}%)` : `hsva(${h}, ${s}%, ${v}%, ${this.roundA})`;
	}
	/**
	* Returns the object as a HSLA object.
	*/
	toHsl() {
		const hsl = rgbToHsl(this.r, this.g, this.b);
		return {
			h: hsl.h * 360,
			s: hsl.s,
			l: hsl.l,
			a: this.a
		};
	}
	/**
	* Returns the hsla values interpolated into a string with the following format:
	* "hsla(xxx, xxx, xxx, xx)".
	*/
	toHslString() {
		const hsl = rgbToHsl(this.r, this.g, this.b);
		const h = Math.round(hsl.h * 360);
		const s = Math.round(hsl.s * 100);
		const l = Math.round(hsl.l * 100);
		return this.a === 1 ? `hsl(${h}, ${s}%, ${l}%)` : `hsla(${h}, ${s}%, ${l}%, ${this.roundA})`;
	}
	/**
	* Returns the hex value of the color.
	* @param allow3Char will shorten hex value to 3 char if possible
	*/
	toHex(allow3Char = false) {
		return rgbToHex(this.r, this.g, this.b, allow3Char);
	}
	/**
	* Returns the hex value of the color -with a # prefixed.
	* @param allow3Char will shorten hex value to 3 char if possible
	*/
	toHexString(allow3Char = false) {
		return "#" + this.toHex(allow3Char);
	}
	/**
	* Returns the hex 8 value of the color.
	* @param allow4Char will shorten hex value to 4 char if possible
	*/
	toHex8(allow4Char = false) {
		return rgbaToHex(this.r, this.g, this.b, this.a, allow4Char);
	}
	/**
	* Returns the hex 8 value of the color -with a # prefixed.
	* @param allow4Char will shorten hex value to 4 char if possible
	*/
	toHex8String(allow4Char = false) {
		return "#" + this.toHex8(allow4Char);
	}
	/**
	* Returns the shorter hex value of the color depends on its alpha -with a # prefixed.
	* @param allowShortChar will shorten hex value to 3 or 4 char if possible
	*/
	toHexShortString(allowShortChar = false) {
		return this.a === 1 ? this.toHexString(allowShortChar) : this.toHex8String(allowShortChar);
	}
	/**
	* Returns the object as a RGBA object.
	*/
	toRgb() {
		return {
			r: Math.round(this.r),
			g: Math.round(this.g),
			b: Math.round(this.b),
			a: this.a
		};
	}
	/**
	* Returns the RGBA values interpolated into a string with the following format:
	* "RGBA(xxx, xxx, xxx, xx)".
	*/
	toRgbString() {
		const r = Math.round(this.r);
		const g = Math.round(this.g);
		const b = Math.round(this.b);
		return this.a === 1 ? `rgb(${r}, ${g}, ${b})` : `rgba(${r}, ${g}, ${b}, ${this.roundA})`;
	}
	/**
	* Returns the object as a RGBA object.
	*/
	toPercentageRgb() {
		const fmt = (x) => `${Math.round(bound01(x, 255) * 100)}%`;
		return {
			r: fmt(this.r),
			g: fmt(this.g),
			b: fmt(this.b),
			a: this.a
		};
	}
	/**
	* Returns the RGBA relative values interpolated into a string
	*/
	toPercentageRgbString() {
		const rnd = (x) => Math.round(bound01(x, 255) * 100);
		return this.a === 1 ? `rgb(${rnd(this.r)}%, ${rnd(this.g)}%, ${rnd(this.b)}%)` : `rgba(${rnd(this.r)}%, ${rnd(this.g)}%, ${rnd(this.b)}%, ${this.roundA})`;
	}
	toCmyk() {
		return { ...rgbToCmyk(this.r, this.g, this.b) };
	}
	toCmykString() {
		const { c, m, y, k } = rgbToCmyk(this.r, this.g, this.b);
		return `cmyk(${c}, ${m}, ${y}, ${k})`;
	}
	/**
	* The 'real' name of the color -if there is one.
	*/
	toName() {
		if (this.a === 0) return "transparent";
		if (this.a < 1) return false;
		const hex = "#" + rgbToHex(this.r, this.g, this.b, false);
		for (const [key, value] of Object.entries(names)) if (hex === value) return key;
		return false;
	}
	toString(format) {
		const formatSet = Boolean(format);
		format = format ?? this.format;
		let formattedString = false;
		const hasAlpha = this.a < 1 && this.a >= 0;
		if (!formatSet && hasAlpha && (format.startsWith("hex") || format === "name")) {
			if (format === "name" && this.a === 0) return this.toName();
			return this.toRgbString();
		}
		if (format === "rgb") formattedString = this.toRgbString();
		if (format === "prgb") formattedString = this.toPercentageRgbString();
		if (format === "hex" || format === "hex6") formattedString = this.toHexString();
		if (format === "hex3") formattedString = this.toHexString(true);
		if (format === "hex4") formattedString = this.toHex8String(true);
		if (format === "hex8") formattedString = this.toHex8String();
		if (format === "name") formattedString = this.toName();
		if (format === "hsl") formattedString = this.toHslString();
		if (format === "hsv") formattedString = this.toHsvString();
		if (format === "cmyk") formattedString = this.toCmykString();
		return formattedString || this.toHexString();
	}
	toNumber() {
		return (Math.round(this.r) << 16) + (Math.round(this.g) << 8) + Math.round(this.b);
	}
	clone() {
		return new TinyColor(this.toString());
	}
	/**
	* Lighten the color a given amount. Providing 100 will always return white.
	* @param amount - valid between 1-100
	*/
	lighten(amount = 10) {
		const hsl = this.toHsl();
		hsl.l += amount / 100;
		hsl.l = clamp01(hsl.l);
		return new TinyColor(hsl);
	}
	/**
	* Brighten the color a given amount, from 0 to 100.
	* @param amount - valid between 1-100
	*/
	brighten(amount = 10) {
		const rgb = this.toRgb();
		rgb.r = Math.max(0, Math.min(255, rgb.r - Math.round(255 * -(amount / 100))));
		rgb.g = Math.max(0, Math.min(255, rgb.g - Math.round(255 * -(amount / 100))));
		rgb.b = Math.max(0, Math.min(255, rgb.b - Math.round(255 * -(amount / 100))));
		return new TinyColor(rgb);
	}
	/**
	* Darken the color a given amount, from 0 to 100.
	* Providing 100 will always return black.
	* @param amount - valid between 1-100
	*/
	darken(amount = 10) {
		const hsl = this.toHsl();
		hsl.l -= amount / 100;
		hsl.l = clamp01(hsl.l);
		return new TinyColor(hsl);
	}
	/**
	* Mix the color with pure white, from 0 to 100.
	* Providing 0 will do nothing, providing 100 will always return white.
	* @param amount - valid between 1-100
	*/
	tint(amount = 10) {
		return this.mix("white", amount);
	}
	/**
	* Mix the color with pure black, from 0 to 100.
	* Providing 0 will do nothing, providing 100 will always return black.
	* @param amount - valid between 1-100
	*/
	shade(amount = 10) {
		return this.mix("black", amount);
	}
	/**
	* Desaturate the color a given amount, from 0 to 100.
	* Providing 100 will is the same as calling greyscale
	* @param amount - valid between 1-100
	*/
	desaturate(amount = 10) {
		const hsl = this.toHsl();
		hsl.s -= amount / 100;
		hsl.s = clamp01(hsl.s);
		return new TinyColor(hsl);
	}
	/**
	* Saturate the color a given amount, from 0 to 100.
	* @param amount - valid between 1-100
	*/
	saturate(amount = 10) {
		const hsl = this.toHsl();
		hsl.s += amount / 100;
		hsl.s = clamp01(hsl.s);
		return new TinyColor(hsl);
	}
	/**
	* Completely desaturates a color into greyscale.
	* Same as calling `desaturate(100)`
	*/
	greyscale() {
		return this.desaturate(100);
	}
	/**
	* Spin takes a positive or negative amount within [-360, 360] indicating the change of hue.
	* Values outside of this range will be wrapped into this range.
	*/
	spin(amount) {
		const hsl = this.toHsl();
		const hue = (hsl.h + amount) % 360;
		hsl.h = hue < 0 ? 360 + hue : hue;
		return new TinyColor(hsl);
	}
	/**
	* Mix the current color a given amount with another color, from 0 to 100.
	* 0 means no mixing (return current color).
	*/
	mix(color, amount = 50) {
		const rgb1 = this.toRgb();
		const rgb2 = new TinyColor(color).toRgb();
		const p = amount / 100;
		const rgba = {
			r: (rgb2.r - rgb1.r) * p + rgb1.r,
			g: (rgb2.g - rgb1.g) * p + rgb1.g,
			b: (rgb2.b - rgb1.b) * p + rgb1.b,
			a: (rgb2.a - rgb1.a) * p + rgb1.a
		};
		return new TinyColor(rgba);
	}
	analogous(results = 6, slices = 30) {
		const hsl = this.toHsl();
		const part = 360 / slices;
		const ret = [this];
		for (hsl.h = (hsl.h - (part * results >> 1) + 720) % 360; --results;) {
			hsl.h = (hsl.h + part) % 360;
			ret.push(new TinyColor(hsl));
		}
		return ret;
	}
	/**
	* taken from https://github.com/infusion/jQuery-xcolor/blob/master/jquery.xcolor.js
	*/
	complement() {
		const hsl = this.toHsl();
		hsl.h = (hsl.h + 180) % 360;
		return new TinyColor(hsl);
	}
	monochromatic(results = 6) {
		const hsv = this.toHsv();
		const { h } = hsv;
		const { s } = hsv;
		let { v } = hsv;
		const res = [];
		const modification = 1 / results;
		while (results--) {
			res.push(new TinyColor({
				h,
				s,
				v
			}));
			v = (v + modification) % 1;
		}
		return res;
	}
	splitcomplement() {
		const hsl = this.toHsl();
		const { h } = hsl;
		return [
			this,
			new TinyColor({
				h: (h + 72) % 360,
				s: hsl.s,
				l: hsl.l
			}),
			new TinyColor({
				h: (h + 216) % 360,
				s: hsl.s,
				l: hsl.l
			})
		];
	}
	/**
	* Compute how the color would appear on a background
	*/
	onBackground(background) {
		const fg = this.toRgb();
		const bg = new TinyColor(background).toRgb();
		const alpha = fg.a + bg.a * (1 - fg.a);
		return new TinyColor({
			r: (fg.r * fg.a + bg.r * bg.a * (1 - fg.a)) / alpha,
			g: (fg.g * fg.a + bg.g * bg.a * (1 - fg.a)) / alpha,
			b: (fg.b * fg.a + bg.b * bg.a * (1 - fg.a)) / alpha,
			a: alpha
		});
	}
	/**
	* Alias for `polyad(3)`
	*/
	triad() {
		return this.polyad(3);
	}
	/**
	* Alias for `polyad(4)`
	*/
	tetrad() {
		return this.polyad(4);
	}
	/**
	* Get polyad colors, like (for 1, 2, 3, 4, 5, 6, 7, 8, etc...)
	* monad, dyad, triad, tetrad, pentad, hexad, heptad, octad, etc...
	*/
	polyad(n) {
		const hsl = this.toHsl();
		const { h } = hsl;
		const result = [this];
		const increment = 360 / n;
		for (let i = 1; i < n; i++) result.push(new TinyColor({
			h: (h + i * increment) % 360,
			s: hsl.s,
			l: hsl.l
		}));
		return result;
	}
	/**
	* compare color vs current color
	*/
	equals(color) {
		const comparedColor = new TinyColor(color);
		/**
		* RGB and CMYK do not have the same color gamut, so a CMYK conversion will never be 100%.
		* This means we need to compare CMYK to CMYK to ensure accuracy of the equals function.
		*/
		if (this.format === "cmyk" || comparedColor.format === "cmyk") return this.toCmykString() === comparedColor.toCmykString();
		return this.toRgbString() === comparedColor.toRgbString();
	}
};
//#endregion
//#region node_modules/.pnpm/ngx-color@10.1.0_@angular+common@22.2.1_@angular+core@22.2.1_@angular+compiler@22.2.1_r_95d8a4f570d71c262f40cd31b73037e1/node_modules/ngx-color/fesm2022/ngx-color.mjs
var checkboardCache = {};
function render(c1, c2, size) {
	if (typeof document === "undefined") return null;
	const canvas = document.createElement("canvas");
	canvas.width = size * 2;
	canvas.height = size * 2;
	const ctx = canvas.getContext("2d");
	if (!ctx) return null;
	ctx.fillStyle = c1;
	ctx.fillRect(0, 0, canvas.width, canvas.height);
	ctx.fillStyle = c2;
	ctx.fillRect(0, 0, size, size);
	ctx.translate(size, size);
	ctx.fillRect(0, 0, size, size);
	return canvas.toDataURL();
}
function getCheckerboard(c1, c2, size) {
	const key = `${c1}-${c2}-${size}`;
	if (checkboardCache[key]) return checkboardCache[key];
	const checkboard = render(c1, c2, size);
	if (!checkboard) return null;
	checkboardCache[key] = checkboard;
	return checkboard;
}
var CheckboardComponent = class CheckboardComponent {
	white = "transparent";
	size = 8;
	grey = "rgba(0,0,0,.08)";
	boxShadow;
	borderRadius;
	gridStyles;
	ngOnInit() {
		const background = getCheckerboard(this.white, this.grey, this.size);
		this.gridStyles = {
			borderRadius: this.borderRadius,
			boxShadow: this.boxShadow,
			background: `url(${background}) center left`
		};
	}
	static ɵfac = function CheckboardComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || CheckboardComponent)();
	};
	static ɵcmp = /*@__PURE__*/ ɵɵdefineComponent({
		type: CheckboardComponent,
		selectors: [["color-checkboard"]],
		inputs: {
			white: "white",
			size: "size",
			grey: "grey",
			boxShadow: "boxShadow",
			borderRadius: "borderRadius"
		},
		standalone: false,
		decls: 1,
		vars: 1,
		consts: [[
			1,
			"grid",
			3,
			"ngStyle"
		]],
		template: function CheckboardComponent_Template(rf, ctx) {
			if (rf & 1) ɵɵelement(0, "div", 0);
			if (rf & 2) ɵɵproperty("ngStyle", ctx.gridStyles);
		},
		dependencies: [NgStyle],
		styles: [".grid[_ngcontent-%COMP%]{inset:0;position:absolute}"]
	});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(CheckboardComponent, [{
		type: Component,
		args: [{
			selector: "color-checkboard",
			template: `<div class="grid" [ngStyle]="gridStyles"></div>`,
			preserveWhitespaces: false,
			changeDetection: ChangeDetectionStrategy.OnPush,
			standalone: false,
			styles: [".grid{inset:0;position:absolute}\n"]
		}]
	}], null, {
		white: [{ type: Input }],
		size: [{ type: Input }],
		grey: [{ type: Input }],
		boxShadow: [{ type: Input }],
		borderRadius: [{ type: Input }]
	});
})();
var CheckboardModule = class CheckboardModule {
	static ɵfac = function CheckboardModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || CheckboardModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: CheckboardModule,
		declarations: [CheckboardComponent],
		imports: [CommonModule],
		exports: [CheckboardComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(CheckboardModule, [{
		type: NgModule,
		args: [{
			declarations: [CheckboardComponent],
			exports: [CheckboardComponent],
			imports: [CommonModule]
		}]
	}], null, null);
})();
var CoordinatesDirective = class CoordinatesDirective {
	el = inject(ElementRef);
	coordinatesChange = new Subject();
	mousechange = new Subject();
	mouseListening = false;
	sub;
	mousemove($event, x, y, isTouch = false) {
		if (this.mouseListening) {
			$event.preventDefault();
			this.mousechange.next({
				$event,
				x,
				y,
				isTouch
			});
		}
	}
	mouseup() {
		this.mouseListening = false;
	}
	mousedown($event, x, y, isTouch = false) {
		$event.preventDefault();
		this.mouseListening = true;
		this.mousechange.next({
			$event,
			x,
			y,
			isTouch
		});
	}
	ngOnInit() {
		this.sub = this.mousechange.pipe(distinctUntilChanged((p, q) => p.x === q.x && p.y === q.y)).subscribe((n) => this.handleChange(n.x, n.y, n.$event, n.isTouch));
	}
	ngOnDestroy() {
		this.sub?.unsubscribe();
	}
	handleChange(x, y, $event, isTouch) {
		const containerWidth = this.el.nativeElement.clientWidth;
		const containerHeight = this.el.nativeElement.clientHeight;
		const left = x - (this.el.nativeElement.getBoundingClientRect().left + window.pageXOffset);
		let top = y - this.el.nativeElement.getBoundingClientRect().top;
		if (!isTouch) top = top - window.pageYOffset;
		this.coordinatesChange.next({
			x,
			y,
			top,
			left,
			containerWidth,
			containerHeight,
			$event
		});
	}
	static ɵfac = function CoordinatesDirective_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || CoordinatesDirective)();
	};
	static ɵdir = /*@__PURE__*/ ɵɵdefineDirective({
		type: CoordinatesDirective,
		selectors: [[
			"",
			"ngx-color-coordinates",
			""
		]],
		hostBindings: function CoordinatesDirective_HostBindings(rf, ctx) {
			if (rf & 1) ɵɵlistener("mousemove", function CoordinatesDirective_mousemove_HostBindingHandler($event) {
				return ctx.mousemove($event, $event.pageX, $event.pageY);
			}, ɵɵresolveWindow)("touchmove", function CoordinatesDirective_touchmove_HostBindingHandler($event) {
				return ctx.mousemove($event, $event.touches[0].clientX, $event.touches[0].clientY, true);
			}, ɵɵresolveWindow)("mouseup", function CoordinatesDirective_mouseup_HostBindingHandler() {
				return ctx.mouseup();
			}, ɵɵresolveWindow)("touchend", function CoordinatesDirective_touchend_HostBindingHandler() {
				return ctx.mouseup();
			}, ɵɵresolveWindow)("mousedown", function CoordinatesDirective_mousedown_HostBindingHandler($event) {
				return ctx.mousedown($event, $event.pageX, $event.pageY);
			})("touchstart", function CoordinatesDirective_touchstart_HostBindingHandler($event) {
				return ctx.mousedown($event, $event.touches[0].clientX, $event.touches[0].clientY, true);
			});
		},
		outputs: { coordinatesChange: "coordinatesChange" },
		standalone: false
	});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(CoordinatesDirective, [{
		type: Directive,
		args: [{
			selector: "[ngx-color-coordinates]",
			standalone: false
		}]
	}], null, {
		coordinatesChange: [{ type: Output }],
		mousemove: [{
			type: HostListener,
			args: ["window:mousemove", [
				"$event",
				"$event.pageX",
				"$event.pageY"
			]]
		}, {
			type: HostListener,
			args: ["window:touchmove", [
				"$event",
				"$event.touches[0].clientX",
				"$event.touches[0].clientY",
				"true"
			]]
		}],
		mouseup: [{
			type: HostListener,
			args: ["window:mouseup"]
		}, {
			type: HostListener,
			args: ["window:touchend"]
		}],
		mousedown: [{
			type: HostListener,
			args: ["mousedown", [
				"$event",
				"$event.pageX",
				"$event.pageY"
			]]
		}, {
			type: HostListener,
			args: ["touchstart", [
				"$event",
				"$event.touches[0].clientX",
				"$event.touches[0].clientY",
				"true"
			]]
		}]
	});
})();
var CoordinatesModule = class CoordinatesModule {
	static ɵfac = function CoordinatesModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || CoordinatesModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: CoordinatesModule,
		declarations: [CoordinatesDirective],
		exports: [CoordinatesDirective]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(CoordinatesModule, [{
		type: NgModule,
		args: [{
			declarations: [CoordinatesDirective],
			exports: [CoordinatesDirective]
		}]
	}], null, null);
})();
var AlphaComponent = class AlphaComponent {
	hsl;
	rgb;
	pointer;
	shadow;
	radius;
	direction = "horizontal";
	onChange = new EventEmitter();
	gradient;
	pointerLeft;
	pointerTop;
	ngOnChanges() {
		if (this.direction === "vertical") {
			this.pointerLeft = 0;
			this.pointerTop = this.rgb.a * 100;
			this.gradient = { background: `linear-gradient(to bottom, rgba(${this.rgb.r},${this.rgb.g},${this.rgb.b}, 0) 0%,
          rgba(${this.rgb.r},${this.rgb.g},${this.rgb.b}, 1) 100%)` };
		} else {
			this.gradient = { background: `linear-gradient(to right, rgba(${this.rgb.r},${this.rgb.g},${this.rgb.b}, 0) 0%,
          rgba(${this.rgb.r},${this.rgb.g},${this.rgb.b}, 1) 100%)` };
			this.pointerLeft = this.rgb.a * 100;
		}
	}
	handleChange({ top, left, containerHeight, containerWidth, $event }) {
		let data;
		if (this.direction === "vertical") {
			let a;
			if (top < 0) a = 0;
			else if (top > containerHeight) a = 1;
			else a = Math.round(top * 100 / containerHeight) / 100;
			if (this.hsl.a !== a) data = {
				h: this.hsl.h,
				s: this.hsl.s,
				l: this.hsl.l,
				a,
				source: "rgb"
			};
		} else {
			let a;
			if (left < 0) a = 0;
			else if (left > containerWidth) a = 1;
			else a = Math.round(left * 100 / containerWidth) / 100;
			if (this.hsl.a !== a) data = {
				h: this.hsl.h,
				s: this.hsl.s,
				l: this.hsl.l,
				a,
				source: "rgb"
			};
		}
		if (!data) return;
		this.onChange.emit({
			data,
			$event
		});
	}
	static ɵfac = function AlphaComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || AlphaComponent)();
	};
	static ɵcmp = /*@__PURE__*/ ɵɵdefineComponent({
		type: AlphaComponent,
		selectors: [["color-alpha"]],
		inputs: {
			hsl: "hsl",
			rgb: "rgb",
			pointer: "pointer",
			shadow: "shadow",
			radius: "radius",
			direction: "direction"
		},
		outputs: { onChange: "onChange" },
		standalone: false,
		features: [ɵɵNgOnChangesFeature],
		decls: 7,
		vars: 15,
		consts: [
			[1, "alpha"],
			[1, "alpha-checkboard"],
			[
				1,
				"alpha-gradient",
				3,
				"ngStyle"
			],
			[
				"ngx-color-coordinates",
				"",
				3,
				"coordinatesChange"
			],
			[1, "alpha-pointer"],
			[
				1,
				"alpha-slider",
				3,
				"ngStyle"
			]
		],
		template: function AlphaComponent_Template(rf, ctx) {
			if (rf & 1) {
				ɵɵelementStart(0, "div", 0)(1, "div", 1);
				ɵɵelement(2, "color-checkboard");
				ɵɵelementEnd();
				ɵɵelement(3, "div", 2);
				ɵɵelementStart(4, "div", 3);
				ɵɵlistener("coordinatesChange", function AlphaComponent_Template_div_coordinatesChange_4_listener($event) {
					return ctx.handleChange($event);
				});
				ɵɵelementStart(5, "div", 4);
				ɵɵelement(6, "div", 5);
				ɵɵelementEnd()()();
			}
			if (rf & 2) {
				ɵɵstyleProp("border-radius", ctx.radius);
				ɵɵadvance(3);
				ɵɵstyleProp("box-shadow", ctx.shadow)("border-radius", ctx.radius);
				ɵɵproperty("ngStyle", ctx.gradient);
				ɵɵadvance();
				ɵɵclassMap(ɵɵinterpolate1("alpha-container color-alpha-", ctx.direction));
				ɵɵadvance();
				ɵɵstyleProp("left", ctx.pointerLeft, "%")("top", ctx.pointerTop, "%");
				ɵɵadvance();
				ɵɵproperty("ngStyle", ctx.pointer);
			}
		},
		dependencies: [
			NgStyle,
			CheckboardComponent,
			CoordinatesDirective
		],
		styles: [".alpha[_ngcontent-%COMP%]{position:absolute;inset:0}.alpha-checkboard[_ngcontent-%COMP%]{position:absolute;inset:0;overflow:hidden}.alpha-gradient[_ngcontent-%COMP%]{position:absolute;inset:0}.alpha-container[_ngcontent-%COMP%]{position:relative;height:100%;margin:0 3px}.alpha-pointer[_ngcontent-%COMP%]{position:absolute}.alpha-slider[_ngcontent-%COMP%]{width:4px;border-radius:1px;height:8px;box-shadow:0 0 2px #0009;background:#fff;margin-top:1px;transform:translate(-2px)}"]
	});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(AlphaComponent, [{
		type: Component,
		args: [{
			selector: "color-alpha",
			template: `
    <div class="alpha" [style.border-radius]="radius">
      <div class="alpha-checkboard">
        <color-checkboard></color-checkboard>
      </div>
      <div
        class="alpha-gradient"
        [ngStyle]="gradient"
        [style.box-shadow]="shadow"
        [style.border-radius]="radius"
      ></div>
      <div
        ngx-color-coordinates
        (coordinatesChange)="handleChange($event)"
        class="alpha-container color-alpha-{{ direction }}"
      >
        <div class="alpha-pointer" [style.left.%]="pointerLeft" [style.top.%]="pointerTop">
          <div class="alpha-slider" [ngStyle]="pointer"></div>
        </div>
      </div>
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			preserveWhitespaces: false,
			standalone: false,
			styles: [".alpha{position:absolute;inset:0}.alpha-checkboard{position:absolute;inset:0;overflow:hidden}.alpha-gradient{position:absolute;inset:0}.alpha-container{position:relative;height:100%;margin:0 3px}.alpha-pointer{position:absolute}.alpha-slider{width:4px;border-radius:1px;height:8px;box-shadow:0 0 2px #0009;background:#fff;margin-top:1px;transform:translate(-2px)}\n"]
		}]
	}], null, {
		hsl: [{ type: Input }],
		rgb: [{ type: Input }],
		pointer: [{ type: Input }],
		shadow: [{ type: Input }],
		radius: [{ type: Input }],
		direction: [{ type: Input }],
		onChange: [{ type: Output }]
	});
})();
var AlphaModule = class AlphaModule {
	static ɵfac = function AlphaModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || AlphaModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: AlphaModule,
		declarations: [AlphaComponent],
		imports: [
			CommonModule,
			CheckboardModule,
			CoordinatesModule
		],
		exports: [AlphaComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [
		CommonModule,
		CheckboardModule,
		CoordinatesModule
	] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(AlphaModule, [{
		type: NgModule,
		args: [{
			declarations: [AlphaComponent],
			exports: [AlphaComponent],
			imports: [
				CommonModule,
				CheckboardModule,
				CoordinatesModule
			]
		}]
	}], null, null);
})();
function simpleCheckForValidColor(data) {
	const keysToCheck = [
		"r",
		"g",
		"b",
		"a",
		"h",
		"s",
		"l",
		"v"
	];
	let checked = 0;
	let passed = 0;
	keysToCheck.forEach((letter) => {
		if (!data[letter]) return;
		checked += 1;
		if (!isNaN(data[letter])) passed += 1;
		if (letter === "s" || letter === "l") {
			if (/^\d+%$/.test(data[letter])) passed += 1;
		}
	});
	return checked === passed ? data : false;
}
function toState(data, oldHue, disableAlpha) {
	const color = data.hex ? new TinyColor(data.hex) : new TinyColor(data);
	if (disableAlpha) color.setAlpha(1);
	const hsl = color.toHsl();
	const hsv = color.toHsv();
	const rgb = color.toRgb();
	const hex = color.toHex();
	if (hsl.s === 0) {
		hsl.h = oldHue || 0;
		hsv.h = oldHue || 0;
	}
	return {
		hsl,
		hex: hex === "000000" && rgb.a === 0 ? "transparent" : color.toHexString(),
		rgb,
		hsv,
		oldHue: data.h || oldHue || hsl.h,
		source: data.source
	};
}
function isValidHex(hex) {
	return new TinyColor(hex).isValid;
}
var ColorMode;
(function(ColorMode) {
	ColorMode["HEX"] = "hex";
	ColorMode["HSL"] = "hsl";
	ColorMode["HSV"] = "hsv";
	ColorMode["RGB"] = "rgb";
})(ColorMode || (ColorMode = {}));
var ColorWrap = class ColorWrap {
	className;
	/**
	* Descriptors the return color format if the component is used with two-way binding
	*/
	mode = ColorMode.HEX;
	color = {
		h: 250,
		s: .5,
		l: .2,
		a: 1
	};
	colorChange = new EventEmitter();
	onChange = new EventEmitter();
	onChangeComplete = new EventEmitter();
	onSwatchHover = new EventEmitter();
	oldHue;
	hsl;
	hsv;
	rgb;
	hex;
	source;
	currentColor;
	changes;
	disableAlpha;
	_onChangeCompleteSubscription = new Subscription();
	_onSwatchHoverSubscription = new Subscription();
	ngOnInit() {
		this.changes = this.onChange.pipe(debounceTime(100), tap((event) => {
			this.onChangeComplete.emit(event);
			switch (this.mode) {
				case ColorMode.HEX:
					this.colorChange.emit(event.color.hex);
					break;
				case ColorMode.HSL:
					this.colorChange.emit(event.color.hsl);
					break;
				case ColorMode.HSV:
					this.colorChange.emit(event.color.hsv);
					break;
				case ColorMode.RGB:
					this.colorChange.emit(event.color.rgb);
					break;
				default:
					const msg = `The mode '${this.mode}' is not supported`;
					if (isDevMode()) throw new Error(msg);
					else console.warn(msg);
			}
		})).subscribe();
		this.setState(toState(this.color, 0));
		this.currentColor = this.hex;
	}
	ngOnChanges() {
		this.setState(toState(this.color, this.oldHue));
	}
	ngOnDestroy() {
		this.changes?.unsubscribe();
		this._onChangeCompleteSubscription?.unsubscribe();
		this._onSwatchHoverSubscription?.unsubscribe();
	}
	setState(data) {
		this.oldHue = data.oldHue;
		this.hsl = data.hsl;
		this.hsv = data.hsv;
		this.rgb = data.rgb;
		this.hex = data.hex;
		this.source = data.source;
		this.afterValidChange();
	}
	handleChange(data, $event) {
		if (simpleCheckForValidColor(data)) {
			const color = toState(data, data.h || this.oldHue, this.disableAlpha);
			this.setState(color);
			this.onChange.emit({
				color,
				$event
			});
			this.afterValidChange();
		}
	}
	/** hook for components after a complete change */
	afterValidChange() {}
	handleSwatchHover(data, $event) {
		if (simpleCheckForValidColor(data)) {
			const color = toState(data, data.h || this.oldHue);
			this.setState(color);
			this.onSwatchHover.emit({
				color,
				$event
			});
		}
	}
	registerOnChange(fn) {
		this._onChangeCompleteSubscription.add(this.onChangeComplete.pipe(tap((event) => fn(event.color.hex))).subscribe());
	}
	registerOnTouched(fn) {
		this._onSwatchHoverSubscription.add(this.onSwatchHover.pipe(tap(() => fn())).subscribe());
	}
	setDisabledState(isDisabled) {}
	writeValue(hex) {
		this.color = hex;
	}
	static ɵfac = function ColorWrap_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || ColorWrap)();
	};
	static ɵcmp = /*@__PURE__*/ ɵɵdefineComponent({
		type: ColorWrap,
		selectors: [["color-wrap"]],
		inputs: {
			className: "className",
			mode: "mode",
			color: "color"
		},
		outputs: {
			colorChange: "colorChange",
			onChange: "onChange",
			onChangeComplete: "onChangeComplete",
			onSwatchHover: "onSwatchHover"
		},
		standalone: false,
		features: [ɵɵProvidersFeature([{
			provide: NG_VALUE_ACCESSOR,
			useExisting: forwardRef(() => ColorWrap),
			multi: true
		}]), ɵɵNgOnChangesFeature],
		decls: 0,
		vars: 0,
		template: function ColorWrap_Template(rf, ctx) {},
		encapsulation: 2,
		changeDetection: 1
	});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(ColorWrap, [{
		type: Component,
		args: [{
			selector: "color-wrap",
			template: ``,
			providers: [{
				provide: NG_VALUE_ACCESSOR,
				useExisting: forwardRef(() => ColorWrap),
				multi: true
			}],
			standalone: false
		}]
	}], null, {
		className: [{ type: Input }],
		mode: [{ type: Input }],
		color: [{ type: Input }],
		colorChange: [{ type: Output }],
		onChange: [{ type: Output }],
		onChangeComplete: [{ type: Output }],
		onSwatchHover: [{ type: Output }]
	});
})();
var ColorWrapModule = class ColorWrapModule {
	static ɵfac = function ColorWrapModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || ColorWrapModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: ColorWrapModule,
		declarations: [ColorWrap],
		imports: [CommonModule],
		exports: [ColorWrap]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(ColorWrapModule, [{
		type: NgModule,
		args: [{
			declarations: [ColorWrap],
			exports: [ColorWrap],
			imports: [CommonModule]
		}]
	}], null, null);
})();
var nextUniqueId = 0;
var EditableInputComponent = class EditableInputComponent {
	style;
	label;
	value;
	arrowOffset;
	dragLabel;
	dragMax;
	placeholder = "";
	onChange = new EventEmitter();
	currentValue;
	blurValue;
	wrapStyle;
	inputStyle;
	labelStyle;
	focus = false;
	mousemove;
	mouseup;
	uniqueId = `editableInput-${++nextUniqueId}`;
	ngOnInit() {
		this.wrapStyle = this.style && this.style.wrap ? this.style.wrap : {};
		this.inputStyle = this.style && this.style.input ? this.style.input : {};
		this.labelStyle = this.style && this.style.label ? this.style.label : {};
		if (this.dragLabel) this.labelStyle.cursor = "ew-resize";
	}
	handleFocus($event) {
		this.focus = true;
	}
	handleFocusOut($event) {
		this.focus = false;
		this.currentValue = this.blurValue;
	}
	handleKeydown($event) {
		const stringValue = String($event.target.value);
		const isPercentage = stringValue.indexOf("%") > -1;
		const num = Number(stringValue.replace(/%/g, ""));
		if (isNaN(num)) return;
		const amount = this.arrowOffset || 1;
		if ($event.keyCode === 38) {
			if (this.label) this.onChange.emit({
				data: { [this.label]: num + amount },
				$event
			});
			else this.onChange.emit({
				data: num + amount,
				$event
			});
			if (isPercentage) this.currentValue = `${num + amount}%`;
			else this.currentValue = num + amount;
		}
		if ($event.keyCode === 40) {
			if (this.label) this.onChange.emit({
				data: { [this.label]: num - amount },
				$event
			});
			else this.onChange.emit({
				data: num - amount,
				$event
			});
			if (isPercentage) this.currentValue = `${num - amount}%`;
			else this.currentValue = num - amount;
		}
	}
	handleKeyup($event) {
		if ($event.keyCode === 40 || $event.keyCode === 38) return;
		if (`${this.currentValue}` === $event.target.value) return;
		if (this.label) this.onChange.emit({
			data: { [this.label]: $event.target.value },
			$event
		});
		else this.onChange.emit({
			data: $event.target.value,
			$event
		});
	}
	ngOnChanges() {
		if (!this.focus) {
			this.currentValue = String(this.value).toUpperCase();
			this.blurValue = String(this.value).toUpperCase();
		} else this.blurValue = String(this.value).toUpperCase();
	}
	ngOnDestroy() {
		this.unsubscribe();
	}
	subscribe() {
		this.mousemove = fromEvent(document, "mousemove").subscribe((ev) => this.handleDrag(ev));
		this.mouseup = fromEvent(document, "mouseup").subscribe(() => this.unsubscribe());
	}
	unsubscribe() {
		this.mousemove?.unsubscribe();
		this.mouseup?.unsubscribe();
	}
	handleMousedown($event) {
		if (this.dragLabel) {
			$event.preventDefault();
			this.handleDrag($event);
			this.subscribe();
		}
	}
	handleDrag($event) {
		if (this.dragLabel) {
			const newValue = Math.round(this.value + $event.movementX);
			if (newValue >= 0 && newValue <= this.dragMax) this.onChange.emit({
				data: { [this.label]: newValue },
				$event
			});
		}
	}
	static ɵfac = function EditableInputComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || EditableInputComponent)();
	};
	static ɵcmp = (function() {
		function EditableInputComponent_Conditional_2_Template(rf, ctx) {
			if (rf & 1) {
				const _r1 = ɵɵgetCurrentView();
				ɵɵelementStart(0, "span", 3);
				ɵɵlistener("mousedown", function EditableInputComponent_Conditional_2_Template_span_mousedown_0_listener($event) {
					ɵɵrestoreView(_r1);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.handleMousedown($event));
				});
				ɵɵtext(1);
				ɵɵelementEnd();
			}
			if (rf & 2) {
				const ctx_r1 = ɵɵnextContext();
				ɵɵproperty("id", ctx_r1.uniqueId)("ngStyle", ctx_r1.labelStyle);
				ɵɵadvance();
				ɵɵtextInterpolate1(" ", ctx_r1.label, " ");
			}
		}
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: EditableInputComponent,
			selectors: [["color-editable-input"]],
			inputs: {
				style: "style",
				label: "label",
				value: "value",
				arrowOffset: "arrowOffset",
				dragLabel: "dragLabel",
				dragMax: "dragMax",
				placeholder: "placeholder"
			},
			outputs: { onChange: "onChange" },
			standalone: false,
			features: [ɵɵNgOnChangesFeature],
			decls: 3,
			vars: 6,
			consts: [
				[
					1,
					"wrap",
					3,
					"ngStyle"
				],
				[
					"spellCheck",
					"false",
					3,
					"keydown",
					"keyup",
					"focus",
					"focusout",
					"ngStyle",
					"value",
					"placeholder"
				],
				[
					3,
					"id",
					"ngStyle"
				],
				[
					3,
					"mousedown",
					"id",
					"ngStyle"
				]
			],
			template: function EditableInputComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵelementStart(0, "div", 0)(1, "input", 1);
					ɵɵlistener("keydown", function EditableInputComponent_Template_input_keydown_1_listener($event) {
						return ctx.handleKeydown($event);
					})("keyup", function EditableInputComponent_Template_input_keyup_1_listener($event) {
						return ctx.handleKeyup($event);
					})("focus", function EditableInputComponent_Template_input_focus_1_listener($event) {
						return ctx.handleFocus($event);
					})("focusout", function EditableInputComponent_Template_input_focusout_1_listener($event) {
						return ctx.handleFocusOut($event);
					});
					ɵɵelementEnd();
					ɵɵconditionalCreate(2, EditableInputComponent_Conditional_2_Template, 2, 3, "span", 2);
					ɵɵelementEnd();
				}
				if (rf & 2) {
					ɵɵproperty("ngStyle", ctx.wrapStyle);
					ɵɵadvance();
					ɵɵproperty("ngStyle", ctx.inputStyle)("value", ctx.currentValue)("placeholder", ctx.placeholder);
					ɵɵattribute("aria-labelledby", ctx.uniqueId);
					ɵɵadvance();
					ɵɵconditional(ctx.label ? 2 : -1);
				}
			},
			dependencies: [NgStyle],
			styles: ["[_nghost-%COMP%]{display:flex}.wrap[_ngcontent-%COMP%]{position:relative}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(EditableInputComponent, [{
		type: Component,
		args: [{
			selector: "color-editable-input",
			template: `
    <div class="wrap" [ngStyle]="wrapStyle">
      <input
        [ngStyle]="inputStyle"
        spellCheck="false"
        [value]="currentValue"
        [placeholder]="placeholder"
        (keydown)="handleKeydown($event)"
        (keyup)="handleKeyup($event)"
        (focus)="handleFocus($event)"
        (focusout)="handleFocusOut($event)"
        [attr.aria-labelledby]="uniqueId"
      />
      @if (label) {
        <span [id]="uniqueId" [ngStyle]="labelStyle" (mousedown)="handleMousedown($event)">
          {{ label }}
        </span>
      }
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			standalone: false,
			styles: [":host{display:flex}.wrap{position:relative}\n"]
		}]
	}], null, {
		style: [{ type: Input }],
		label: [{ type: Input }],
		value: [{ type: Input }],
		arrowOffset: [{ type: Input }],
		dragLabel: [{ type: Input }],
		dragMax: [{ type: Input }],
		placeholder: [{ type: Input }],
		onChange: [{ type: Output }]
	});
})();
var EditableInputModule = class EditableInputModule {
	static ɵfac = function EditableInputModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || EditableInputModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: EditableInputModule,
		declarations: [EditableInputComponent],
		imports: [CommonModule],
		exports: [EditableInputComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(EditableInputModule, [{
		type: NgModule,
		args: [{
			declarations: [EditableInputComponent],
			exports: [EditableInputComponent],
			imports: [CommonModule]
		}]
	}], null, null);
})();
var HueComponent = class HueComponent {
	hsl;
	pointer;
	radius;
	shadow;
	hidePointer = false;
	direction = "horizontal";
	onChange = new EventEmitter();
	left = "0px";
	top = "";
	ngOnChanges() {
		if (this.direction === "horizontal") this.left = `${this.hsl.h * 100 / 360}%`;
		else this.top = `${-(this.hsl.h * 100 / 360) + 100}%`;
	}
	handleChange({ top, left, containerHeight, containerWidth, $event }) {
		let data;
		if (this.direction === "vertical") {
			let h;
			if (top < 0) h = 359;
			else if (top > containerHeight) h = 0;
			else h = 360 * (-(top * 100 / containerHeight) + 100) / 100;
			if (this.hsl.h !== h) data = {
				h,
				s: this.hsl.s,
				l: this.hsl.l,
				a: this.hsl.a,
				source: "rgb"
			};
		} else {
			let h;
			if (left < 0) h = 0;
			else if (left > containerWidth) h = 359;
			else h = 360 * (left * 100 / containerWidth) / 100;
			if (this.hsl.h !== h) data = {
				h,
				s: this.hsl.s,
				l: this.hsl.l,
				a: this.hsl.a,
				source: "rgb"
			};
		}
		if (!data) return;
		this.onChange.emit({
			data,
			$event
		});
	}
	static ɵfac = function HueComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || HueComponent)();
	};
	static ɵcmp = (function() {
		function HueComponent_Conditional_2_Template(rf, ctx) {
			if (rf & 1) {
				ɵɵelementStart(0, "div", 2);
				ɵɵelement(1, "div", 3);
				ɵɵelementEnd();
			}
			if (rf & 2) {
				const ctx_r0 = ɵɵnextContext();
				ɵɵstyleProp("left", ctx_r0.left)("top", ctx_r0.top);
				ɵɵadvance();
				ɵɵproperty("ngStyle", ctx_r0.pointer);
			}
		}
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: HueComponent,
			selectors: [["color-hue"]],
			inputs: {
				hsl: "hsl",
				pointer: "pointer",
				radius: "radius",
				shadow: "shadow",
				hidePointer: "hidePointer",
				direction: "direction"
			},
			outputs: { onChange: "onChange" },
			standalone: false,
			features: [ɵɵNgOnChangesFeature],
			decls: 3,
			vars: 8,
			consts: [
				[
					"ngx-color-coordinates",
					"",
					1,
					"color-hue-container",
					3,
					"coordinatesChange"
				],
				[
					1,
					"color-hue-pointer",
					3,
					"left",
					"top"
				],
				[1, "color-hue-pointer"],
				[
					1,
					"color-hue-slider",
					3,
					"ngStyle"
				]
			],
			template: function HueComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵelementStart(0, "div")(1, "div", 0);
					ɵɵlistener("coordinatesChange", function HueComponent_Template_div_coordinatesChange_1_listener($event) {
						return ctx.handleChange($event);
					});
					ɵɵconditionalCreate(2, HueComponent_Conditional_2_Template, 2, 5, "div", 1);
					ɵɵelementEnd()();
				}
				if (rf & 2) {
					ɵɵclassMap(ɵɵinterpolate1("color-hue color-hue-", ctx.direction));
					ɵɵstyleProp("border-radius", ctx.radius, "px")("box-shadow", ctx.shadow);
					ɵɵadvance(2);
					ɵɵconditional(!ctx.hidePointer ? 2 : -1);
				}
			},
			dependencies: [NgStyle, CoordinatesDirective],
			styles: [".color-hue[_ngcontent-%COMP%]{position:absolute;inset:0}.color-hue-container[_ngcontent-%COMP%]{margin:0 2px;position:relative;height:100%}.color-hue-pointer[_ngcontent-%COMP%]{position:absolute}.color-hue-slider[_ngcontent-%COMP%]{margin-top:1px;width:4px;border-radius:1px;height:8px;box-shadow:0 0 2px #0009;background:#fff;transform:translate(-2px)}.color-hue-horizontal[_ngcontent-%COMP%]{background:linear-gradient(to right,red,#ff0 17%,#0f0 33%,#0ff,#00f 67%,#f0f 83%,red)}.color-hue-vertical[_ngcontent-%COMP%]{background:linear-gradient(to top,red,#ff0 17%,#0f0 33%,#0ff,#00f 67%,#f0f 83%,red)}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(HueComponent, [{
		type: Component,
		args: [{
			selector: "color-hue",
			template: `
    <div
      class="color-hue color-hue-{{ direction }}"
      [style.border-radius.px]="radius"
      [style.box-shadow]="shadow"
    >
      <div
        ngx-color-coordinates
        (coordinatesChange)="handleChange($event)"
        class="color-hue-container"
      >
        @if (!hidePointer) {
          <div class="color-hue-pointer" [style.left]="left" [style.top]="top">
            <div class="color-hue-slider" [ngStyle]="pointer"></div>
          </div>
        }
      </div>
    </div>
  `,
			preserveWhitespaces: false,
			changeDetection: ChangeDetectionStrategy.OnPush,
			standalone: false,
			styles: [".color-hue{position:absolute;inset:0}.color-hue-container{margin:0 2px;position:relative;height:100%}.color-hue-pointer{position:absolute}.color-hue-slider{margin-top:1px;width:4px;border-radius:1px;height:8px;box-shadow:0 0 2px #0009;background:#fff;transform:translate(-2px)}.color-hue-horizontal{background:linear-gradient(to right,red,#ff0 17%,#0f0 33%,#0ff,#00f 67%,#f0f 83%,red)}.color-hue-vertical{background:linear-gradient(to top,red,#ff0 17%,#0f0 33%,#0ff,#00f 67%,#f0f 83%,red)}\n"]
		}]
	}], null, {
		hsl: [{ type: Input }],
		pointer: [{ type: Input }],
		radius: [{ type: Input }],
		shadow: [{ type: Input }],
		hidePointer: [{ type: Input }],
		direction: [{ type: Input }],
		onChange: [{ type: Output }]
	});
})();
var HueModule = class HueModule {
	static ɵfac = function HueModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || HueModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: HueModule,
		declarations: [HueComponent],
		imports: [CommonModule, CoordinatesModule],
		exports: [HueComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule, CoordinatesModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(HueModule, [{
		type: NgModule,
		args: [{
			declarations: [HueComponent],
			exports: [HueComponent],
			imports: [CommonModule, CoordinatesModule]
		}]
	}], null, null);
})();
var RaisedComponent = class RaisedComponent {
	zDepth = 1;
	radius = 1;
	background = "#fff";
	static ɵfac = function RaisedComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || RaisedComponent)();
	};
	static ɵcmp = (function() {
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: RaisedComponent,
			selectors: [["color-raised"]],
			inputs: {
				zDepth: "zDepth",
				radius: "radius",
				background: "background"
			},
			standalone: false,
			ngContentSelectors: ["*"],
			decls: 4,
			vars: 5,
			consts: [[1, "raised-wrap"], [1, "raised-content"]],
			template: function RaisedComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵprojectionDef();
					ɵɵelementStart(0, "div", 0);
					ɵɵelement(1, "div");
					ɵɵelementStart(2, "div", 1);
					ɵɵprojection(3);
					ɵɵelementEnd()();
				}
				if (rf & 2) {
					ɵɵadvance();
					ɵɵclassMap(ɵɵinterpolate1("raised-bg zDepth-", ctx.zDepth));
					ɵɵstyleProp("background", ctx.background);
				}
			},
			styles: [".raised-wrap[_ngcontent-%COMP%]{position:relative;display:inline-block}.raised-bg[_ngcontent-%COMP%]{position:absolute;inset:0}.raised-content[_ngcontent-%COMP%]{position:relative}.zDepth-0[_ngcontent-%COMP%]{box-shadow:none}.zDepth-1[_ngcontent-%COMP%]{box-shadow:0 2px 10px #0000001f,0 2px 5px #00000029}.zDepth-2[_ngcontent-%COMP%]{box-shadow:0 6px 20px #00000030,0 8px 17px #0003}.zDepth-3[_ngcontent-%COMP%]{box-shadow:0 17px 50px #00000030,0 12px 15px #0000003d}.zDepth-4[_ngcontent-%COMP%]{box-shadow:0 25px 55px #00000036,0 16px 28px #00000038}.zDepth-5[_ngcontent-%COMP%]{box-shadow:0 40px 77px #00000038,0 27px 24px #0003}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(RaisedComponent, [{
		type: Component,
		args: [{
			selector: "color-raised",
			template: `
    <div class="raised-wrap">
      <div class="raised-bg zDepth-{{ zDepth }}" [style.background]="background"></div>
      <div class="raised-content">
        <ng-content></ng-content>
      </div>
    </div>
  `,
			preserveWhitespaces: false,
			changeDetection: ChangeDetectionStrategy.OnPush,
			standalone: false,
			styles: [".raised-wrap{position:relative;display:inline-block}.raised-bg{position:absolute;inset:0}.raised-content{position:relative}.zDepth-0{box-shadow:none}.zDepth-1{box-shadow:0 2px 10px #0000001f,0 2px 5px #00000029}.zDepth-2{box-shadow:0 6px 20px #00000030,0 8px 17px #0003}.zDepth-3{box-shadow:0 17px 50px #00000030,0 12px 15px #0000003d}.zDepth-4{box-shadow:0 25px 55px #00000036,0 16px 28px #00000038}.zDepth-5{box-shadow:0 40px 77px #00000038,0 27px 24px #0003}\n"]
		}]
	}], null, {
		zDepth: [{ type: Input }],
		radius: [{ type: Input }],
		background: [{ type: Input }]
	});
})();
var RaisedModule = class RaisedModule {
	static ɵfac = function RaisedModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || RaisedModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: RaisedModule,
		declarations: [RaisedComponent],
		imports: [CommonModule],
		exports: [RaisedComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(RaisedModule, [{
		type: NgModule,
		args: [{
			declarations: [RaisedComponent],
			exports: [RaisedComponent],
			imports: [CommonModule]
		}]
	}], null, null);
})();
var SaturationComponent = class SaturationComponent {
	hsl;
	hsv;
	radius;
	pointer;
	circle;
	onChange = new EventEmitter();
	background;
	pointerTop;
	pointerLeft;
	ngOnChanges() {
		this.background = `hsl(${this.hsl.h}, 100%, 50%)`;
		this.pointerTop = -(this.hsv.v * 100) + 1 + 100 + "%";
		this.pointerLeft = this.hsv.s * 100 + "%";
	}
	handleChange({ top, left, containerHeight, containerWidth, $event }) {
		if (left < 0) left = 0;
		else if (left > containerWidth) left = containerWidth;
		else if (top < 0) top = 0;
		else if (top > containerHeight) top = containerHeight;
		const saturation = left / containerWidth;
		let bright = -(top / containerHeight) + 1;
		bright = bright > 0 ? bright : 0;
		bright = bright > 1 ? 1 : bright;
		const data = {
			h: this.hsl.h,
			s: saturation,
			v: bright,
			a: this.hsl.a,
			source: "hsva"
		};
		this.onChange.emit({
			data,
			$event
		});
	}
	static ɵfac = function SaturationComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SaturationComponent)();
	};
	static ɵcmp = /*@__PURE__*/ ɵɵdefineComponent({
		type: SaturationComponent,
		selectors: [["color-saturation"]],
		inputs: {
			hsl: "hsl",
			hsv: "hsv",
			radius: "radius",
			pointer: "pointer",
			circle: "circle"
		},
		outputs: { onChange: "onChange" },
		standalone: false,
		features: [ɵɵNgOnChangesFeature],
		decls: 5,
		vars: 8,
		consts: [
			[
				"ngx-color-coordinates",
				"",
				1,
				"color-saturation",
				3,
				"coordinatesChange"
			],
			[1, "saturation-white"],
			[1, "saturation-black"],
			[
				1,
				"saturation-pointer",
				3,
				"ngStyle"
			],
			[
				1,
				"saturation-circle",
				3,
				"ngStyle"
			]
		],
		template: function SaturationComponent_Template(rf, ctx) {
			if (rf & 1) {
				ɵɵelementStart(0, "div", 0);
				ɵɵlistener("coordinatesChange", function SaturationComponent_Template_div_coordinatesChange_0_listener($event) {
					return ctx.handleChange($event);
				});
				ɵɵelementStart(1, "div", 1);
				ɵɵelement(2, "div", 2);
				ɵɵelementStart(3, "div", 3);
				ɵɵelement(4, "div", 4);
				ɵɵelementEnd()()();
			}
			if (rf & 2) {
				ɵɵstyleProp("background", ctx.background);
				ɵɵadvance(3);
				ɵɵstyleProp("top", ctx.pointerTop)("left", ctx.pointerLeft);
				ɵɵproperty("ngStyle", ctx.pointer);
				ɵɵadvance();
				ɵɵproperty("ngStyle", ctx.circle);
			}
		},
		dependencies: [NgStyle, CoordinatesDirective],
		styles: [".saturation-white[_ngcontent-%COMP%]{background:linear-gradient(to right,#fff,#fff0);position:absolute;inset:0}.saturation-black[_ngcontent-%COMP%]{background:linear-gradient(to top,#000,#0000);position:absolute;inset:0}.color-saturation[_ngcontent-%COMP%]{position:absolute;inset:0}.saturation-pointer[_ngcontent-%COMP%]{position:absolute;cursor:default}.saturation-circle[_ngcontent-%COMP%]{width:4px;height:4px;box-shadow:0 0 0 1.5px #fff,inset 0 0 1px 1px #0000004d,0 0 1px 2px #0006;border-radius:50%;cursor:hand;transform:translate(-2px,-4px)}"]
	});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SaturationComponent, [{
		type: Component,
		args: [{
			selector: "color-saturation",
			template: `
    <div
      class="color-saturation"
      ngx-color-coordinates
      (coordinatesChange)="handleChange($event)"
      [style.background]="background"
    >
      <div class="saturation-white">
        <div class="saturation-black"></div>
        <div
          class="saturation-pointer"
          [ngStyle]="pointer"
          [style.top]="pointerTop"
          [style.left]="pointerLeft"
        >
          <div class="saturation-circle" [ngStyle]="circle"></div>
        </div>
      </div>
    </div>
  `,
			preserveWhitespaces: false,
			changeDetection: ChangeDetectionStrategy.OnPush,
			standalone: false,
			styles: [".saturation-white{background:linear-gradient(to right,#fff,#fff0);position:absolute;inset:0}.saturation-black{background:linear-gradient(to top,#000,#0000);position:absolute;inset:0}.color-saturation{position:absolute;inset:0}.saturation-pointer{position:absolute;cursor:default}.saturation-circle{width:4px;height:4px;box-shadow:0 0 0 1.5px #fff,inset 0 0 1px 1px #0000004d,0 0 1px 2px #0006;border-radius:50%;cursor:hand;transform:translate(-2px,-4px)}\n"]
		}]
	}], null, {
		hsl: [{ type: Input }],
		hsv: [{ type: Input }],
		radius: [{ type: Input }],
		pointer: [{ type: Input }],
		circle: [{ type: Input }],
		onChange: [{ type: Output }]
	});
})();
var SaturationModule = class SaturationModule {
	static ɵfac = function SaturationModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SaturationModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: SaturationModule,
		declarations: [SaturationComponent],
		imports: [CommonModule, CoordinatesModule],
		exports: [SaturationComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule, CoordinatesModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SaturationModule, [{
		type: NgModule,
		args: [{
			declarations: [SaturationComponent],
			exports: [SaturationComponent],
			imports: [CommonModule, CoordinatesModule]
		}]
	}], null, null);
})();
var SwatchComponent = class SwatchComponent {
	color;
	style = {};
	focusStyle = {};
	focus;
	onClick = new EventEmitter();
	onHover = new EventEmitter();
	divStyles = {};
	focusStyles = {};
	inFocus = false;
	ngOnInit() {
		this.divStyles = {
			background: this.color,
			...this.style
		};
	}
	currentStyles() {
		this.focusStyles = {
			...this.divStyles,
			...this.focusStyle
		};
		return this.focus || this.inFocus ? this.focusStyles : this.divStyles;
	}
	handleFocusOut() {
		this.inFocus = false;
	}
	handleFocus() {
		this.inFocus = true;
	}
	handleHover(hex, $event) {
		this.onHover.emit({
			hex,
			$event
		});
	}
	handleClick(hex, $event) {
		this.onClick.emit({
			hex,
			$event
		});
	}
	static ɵfac = function SwatchComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SwatchComponent)();
	};
	static ɵcmp = (function() {
		const _c0 = ["*"];
		function SwatchComponent_Conditional_2_Template(rf, ctx) {
			if (rf & 1) ɵɵelement(0, "color-checkboard", 1);
		}
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: SwatchComponent,
			selectors: [["color-swatch"]],
			inputs: {
				color: "color",
				style: "style",
				focusStyle: "focusStyle",
				focus: "focus"
			},
			outputs: {
				onClick: "onClick",
				onHover: "onHover"
			},
			standalone: false,
			ngContentSelectors: _c0,
			decls: 3,
			vars: 3,
			consts: [[
				"tabindex",
				"0",
				1,
				"swatch",
				3,
				"click",
				"keydown.enter",
				"focus",
				"blur",
				"mouseover",
				"ngStyle"
			], ["boxShadow", "inset 0 0 0 1px rgba(0,0,0,0.1)"]],
			template: function SwatchComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵprojectionDef();
					ɵɵelementStart(0, "div", 0);
					ɵɵlistener("click", function SwatchComponent_Template_div_click_0_listener($event) {
						return ctx.handleClick(ctx.color, $event);
					})("keydown.enter", function SwatchComponent_Template_div_keydown_enter_0_listener($event) {
						return ctx.handleClick(ctx.color, $event);
					})("focus", function SwatchComponent_Template_div_focus_0_listener() {
						return ctx.handleFocus();
					})("blur", function SwatchComponent_Template_div_blur_0_listener() {
						return ctx.handleFocusOut();
					})("mouseover", function SwatchComponent_Template_div_mouseover_0_listener($event) {
						return ctx.handleHover(ctx.color, $event);
					});
					ɵɵprojection(1);
					ɵɵconditionalCreate(2, SwatchComponent_Conditional_2_Template, 1, 0, "color-checkboard", 1);
					ɵɵelementEnd();
				}
				if (rf & 2) {
					ɵɵproperty("ngStyle", ctx.currentStyles());
					ɵɵattribute("title", ctx.color);
					ɵɵadvance(2);
					ɵɵconditional(ctx.color === "transparent" ? 2 : -1);
				}
			},
			dependencies: [NgStyle, CheckboardComponent],
			styles: [".swatch[_ngcontent-%COMP%]{outline:none;height:100%;width:100%;cursor:pointer;position:relative}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SwatchComponent, [{
		type: Component,
		args: [{
			selector: "color-swatch",
			template: `
    <div
      class="swatch"
      [ngStyle]="currentStyles()"
      [attr.title]="color"
      (click)="handleClick(color, $event)"
      (keydown.enter)="handleClick(color, $event)"
      (focus)="handleFocus()"
      (blur)="handleFocusOut()"
      (mouseover)="handleHover(color, $event)"
      tabindex="0"
    >
      <ng-content></ng-content>
      @if (color === 'transparent') {
        <color-checkboard boxShadow="inset 0 0 0 1px rgba(0,0,0,0.1)"></color-checkboard>
      }
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			standalone: false,
			styles: [".swatch{outline:none;height:100%;width:100%;cursor:pointer;position:relative}\n"]
		}]
	}], null, {
		color: [{ type: Input }],
		style: [{ type: Input }],
		focusStyle: [{ type: Input }],
		focus: [{ type: Input }],
		onClick: [{ type: Output }],
		onHover: [{ type: Output }]
	});
})();
var SwatchModule = class SwatchModule {
	static ɵfac = function SwatchModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SwatchModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: SwatchModule,
		declarations: [SwatchComponent],
		imports: [CommonModule, CheckboardModule],
		exports: [SwatchComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule, CheckboardModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SwatchModule, [{
		type: NgModule,
		args: [{
			declarations: [SwatchComponent],
			exports: [SwatchComponent],
			imports: [CommonModule, CheckboardModule]
		}]
	}], null, null);
})();
var ShadeComponent = class ShadeComponent {
	hsl;
	rgb;
	pointer;
	shadow;
	radius;
	onChange = new EventEmitter();
	gradient;
	pointerLeft;
	pointerTop;
	ngOnChanges() {
		this.gradient = { background: `linear-gradient(to right,
          hsl(${this.hsl.h}, 90%, 55%),
          #000)` };
		const hsv = new TinyColor(this.hsl).toHsv();
		this.pointerLeft = 100 - hsv.v * 100;
	}
	handleChange({ left, containerWidth, $event }) {
		let data;
		let v;
		if (left < 0) v = 0;
		else if (left > containerWidth) v = 1;
		else v = Math.round(left * 100 / containerWidth) / 100;
		if (new TinyColor(this.hsl).toHsv().v !== v) data = {
			h: this.hsl.h,
			s: 100,
			v: 1 - v,
			l: this.hsl.l,
			a: this.hsl.a,
			source: "rgb"
		};
		if (!data) return;
		this.onChange.emit({
			data,
			$event
		});
	}
	static ɵfac = function ShadeComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || ShadeComponent)();
	};
	static ɵcmp = /*@__PURE__*/ ɵɵdefineComponent({
		type: ShadeComponent,
		selectors: [["color-shade"]],
		inputs: {
			hsl: "hsl",
			rgb: "rgb",
			pointer: "pointer",
			shadow: "shadow",
			radius: "radius"
		},
		outputs: { onChange: "onChange" },
		standalone: false,
		features: [ɵɵNgOnChangesFeature],
		decls: 5,
		vars: 12,
		consts: [
			[1, "shade"],
			[
				1,
				"shade-gradient",
				3,
				"ngStyle"
			],
			[
				"ngx-color-coordinates",
				"",
				1,
				"shade-container",
				3,
				"coordinatesChange"
			],
			[1, "shade-pointer"],
			[
				1,
				"shade-slider",
				3,
				"ngStyle"
			]
		],
		template: function ShadeComponent_Template(rf, ctx) {
			if (rf & 1) {
				ɵɵelementStart(0, "div", 0);
				ɵɵelement(1, "div", 1);
				ɵɵelementStart(2, "div", 2);
				ɵɵlistener("coordinatesChange", function ShadeComponent_Template_div_coordinatesChange_2_listener($event) {
					return ctx.handleChange($event);
				});
				ɵɵelementStart(3, "div", 3);
				ɵɵelement(4, "div", 4);
				ɵɵelementEnd()()();
			}
			if (rf & 2) {
				ɵɵstyleProp("border-radius", ctx.radius);
				ɵɵadvance();
				ɵɵstyleProp("box-shadow", ctx.shadow)("border-radius", ctx.radius);
				ɵɵproperty("ngStyle", ctx.gradient);
				ɵɵadvance(2);
				ɵɵstyleProp("left", ctx.pointerLeft, "%")("top", ctx.pointerTop, "%");
				ɵɵadvance();
				ɵɵproperty("ngStyle", ctx.pointer);
			}
		},
		dependencies: [NgStyle, CoordinatesDirective],
		styles: [".shade[_ngcontent-%COMP%], .shade-gradient[_ngcontent-%COMP%]{position:absolute;inset:0}.shade-container[_ngcontent-%COMP%]{position:relative;height:100%;margin:0 3px}.shade-pointer[_ngcontent-%COMP%]{position:absolute}.shade-slider[_ngcontent-%COMP%]{width:4px;border-radius:1px;height:8px;box-shadow:0 0 2px #0009;background:#fff;margin-top:1px;transform:translate(-2px)}"]
	});
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(ShadeComponent, [{
		type: Component,
		args: [{
			selector: "color-shade",
			template: `
    <div class="shade" [style.border-radius]="radius">
      <div
        class="shade-gradient"
        [ngStyle]="gradient"
        [style.box-shadow]="shadow"
        [style.border-radius]="radius"
      ></div>
      <div ngx-color-coordinates (coordinatesChange)="handleChange($event)" class="shade-container">
        <div class="shade-pointer" [style.left.%]="pointerLeft" [style.top.%]="pointerTop">
          <div class="shade-slider" [ngStyle]="pointer"></div>
        </div>
      </div>
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			preserveWhitespaces: false,
			standalone: false,
			styles: [".shade,.shade-gradient{position:absolute;inset:0}.shade-container{position:relative;height:100%;margin:0 3px}.shade-pointer{position:absolute}.shade-slider{width:4px;border-radius:1px;height:8px;box-shadow:0 0 2px #0009;background:#fff;margin-top:1px;transform:translate(-2px)}\n"]
		}]
	}], null, {
		hsl: [{ type: Input }],
		rgb: [{ type: Input }],
		pointer: [{ type: Input }],
		shadow: [{ type: Input }],
		radius: [{ type: Input }],
		onChange: [{ type: Output }]
	});
})();
var ShadeModule = class ShadeModule {
	static ɵfac = function ShadeModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || ShadeModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: ShadeModule,
		declarations: [ShadeComponent],
		imports: [CommonModule, CoordinatesModule],
		exports: [ShadeComponent]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [CommonModule, CoordinatesModule] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(ShadeModule, [{
		type: NgModule,
		args: [{
			declarations: [ShadeComponent],
			exports: [ShadeComponent],
			imports: [CommonModule, CoordinatesModule]
		}]
	}], null, null);
})();
//#endregion
//#region node_modules/.pnpm/ngx-color@10.1.0_@angular+common@22.2.1_@angular+core@22.2.1_@angular+compiler@22.2.1_r_95d8a4f570d71c262f40cd31b73037e1/node_modules/ngx-color/fesm2022/ngx-color-sketch.mjs
var SketchFieldsComponent = class SketchFieldsComponent {
	hsl;
	rgb;
	hex;
	disableAlpha = false;
	onChange = new EventEmitter();
	input = {
		width: "100%",
		padding: "4px 10% 3px",
		border: "none",
		boxSizing: "border-box",
		boxShadow: "inset 0 0 0 1px #ccc",
		fontSize: "11px"
	};
	label = {
		display: "block",
		textAlign: "center",
		fontSize: "11px",
		color: "#222",
		paddingTop: "3px",
		paddingBottom: "4px",
		textTransform: "capitalize"
	};
	round(value) {
		return Math.round(value);
	}
	handleChange({ data, $event }) {
		if (data.hex) {
			if (isValidHex(data.hex)) {
				const color = new TinyColor(data.hex);
				this.onChange.emit({
					data: {
						hex: this.disableAlpha || data.hex.length <= 6 ? color.toHex() : color.toHex8(),
						source: "hex"
					},
					$event
				});
			}
		} else if (data.r || data.g || data.b) this.onChange.emit({
			data: {
				r: data.r || this.rgb.r,
				g: data.g || this.rgb.g,
				b: data.b || this.rgb.b,
				source: "rgb"
			},
			$event
		});
		else if (data.a) {
			if (data.a < 0) data.a = 0;
			else if (data.a > 100) data.a = 100;
			data.a /= 100;
			if (this.disableAlpha) data.a = 1;
			this.onChange.emit({
				data: {
					h: this.hsl.h,
					s: this.hsl.s,
					l: this.hsl.l,
					a: Math.round(data.a * 100) / 100,
					source: "rgb"
				},
				$event
			});
		} else if (data.h || data.s || data.l) this.onChange.emit({
			data: {
				h: data.h || this.hsl.h,
				s: Number(data.s && data.s || this.hsl.s),
				l: Number(data.l && data.l || this.hsl.l),
				source: "hsl"
			},
			$event
		});
	}
	static ɵfac = function SketchFieldsComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SketchFieldsComponent)();
	};
	static ɵcmp = (function() {
		const _c0 = (a0, a1) => ({
			input: a0,
			label: a1
		});
		function SketchFieldsComponent_Conditional_9_Template(rf, ctx) {
			if (rf & 1) {
				const _r1 = ɵɵgetCurrentView();
				ɵɵelementStart(0, "div", 7)(1, "color-editable-input", 8);
				ɵɵlistener("onChange", function SketchFieldsComponent_Conditional_9_Template_color_editable_input_onChange_1_listener($event) {
					ɵɵrestoreView(_r1);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.handleChange($event));
				});
				ɵɵelementEnd()();
			}
			if (rf & 2) {
				const ctx_r1 = ɵɵnextContext();
				ɵɵadvance();
				ɵɵstyleMap(ɵɵpureFunction2(5, _c0, ctx_r1.input, ctx_r1.label));
				ɵɵproperty("value", ctx_r1.round(ctx_r1.rgb.a * 100))("dragLabel", true)("dragMax", 100);
			}
		}
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: SketchFieldsComponent,
			selectors: [["color-sketch-fields"]],
			inputs: {
				hsl: "hsl",
				rgb: "rgb",
				hex: "hex",
				disableAlpha: "disableAlpha"
			},
			outputs: { onChange: "onChange" },
			standalone: false,
			decls: 10,
			vars: 31,
			consts: [
				[1, "sketch-fields"],
				[1, "sketch-double"],
				[
					"label",
					"hex",
					3,
					"onChange",
					"value"
				],
				[1, "sketch-single"],
				[
					"label",
					"r",
					3,
					"onChange",
					"value",
					"dragLabel",
					"dragMax"
				],
				[
					"label",
					"g",
					3,
					"onChange",
					"value",
					"dragLabel",
					"dragMax"
				],
				[
					"label",
					"b",
					3,
					"onChange",
					"value",
					"dragLabel",
					"dragMax"
				],
				[1, "sketch-alpha"],
				[
					"label",
					"a",
					3,
					"onChange",
					"value",
					"dragLabel",
					"dragMax"
				]
			],
			template: function SketchFieldsComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵelementStart(0, "div", 0)(1, "div", 1)(2, "color-editable-input", 2);
					ɵɵlistener("onChange", function SketchFieldsComponent_Template_color_editable_input_onChange_2_listener($event) {
						return ctx.handleChange($event);
					});
					ɵɵelementEnd()();
					ɵɵelementStart(3, "div", 3)(4, "color-editable-input", 4);
					ɵɵlistener("onChange", function SketchFieldsComponent_Template_color_editable_input_onChange_4_listener($event) {
						return ctx.handleChange($event);
					});
					ɵɵelementEnd()();
					ɵɵelementStart(5, "div", 3)(6, "color-editable-input", 5);
					ɵɵlistener("onChange", function SketchFieldsComponent_Template_color_editable_input_onChange_6_listener($event) {
						return ctx.handleChange($event);
					});
					ɵɵelementEnd()();
					ɵɵelementStart(7, "div", 3)(8, "color-editable-input", 6);
					ɵɵlistener("onChange", function SketchFieldsComponent_Template_color_editable_input_onChange_8_listener($event) {
						return ctx.handleChange($event);
					});
					ɵɵelementEnd()();
					ɵɵconditionalCreate(9, SketchFieldsComponent_Conditional_9_Template, 2, 8, "div", 7);
					ɵɵelementEnd();
				}
				if (rf & 2) {
					ɵɵadvance(2);
					ɵɵstyleMap(ɵɵpureFunction2(19, _c0, ctx.input, ctx.label));
					ɵɵproperty("value", ctx.hex.replace("#", ""));
					ɵɵadvance(2);
					ɵɵstyleMap(ɵɵpureFunction2(22, _c0, ctx.input, ctx.label));
					ɵɵproperty("value", ctx.rgb.r)("dragLabel", true)("dragMax", 255);
					ɵɵadvance(2);
					ɵɵstyleMap(ɵɵpureFunction2(25, _c0, ctx.input, ctx.label));
					ɵɵproperty("value", ctx.rgb.g)("dragLabel", true)("dragMax", 255);
					ɵɵadvance(2);
					ɵɵstyleMap(ɵɵpureFunction2(28, _c0, ctx.input, ctx.label));
					ɵɵproperty("value", ctx.rgb.b)("dragLabel", true)("dragMax", 255);
					ɵɵadvance();
					ɵɵconditional(ctx.disableAlpha === false ? 9 : -1);
				}
			},
			dependencies: [EditableInputComponent],
			styles: [".sketch-fields[_ngcontent-%COMP%]{display:flex;padding-top:4px}.sketch-double[_ngcontent-%COMP%]{-webkit-box-flex:2;flex:2 1 0%}.sketch-single[_ngcontent-%COMP%]{flex:1 1 0%;padding-left:6px}.sketch-alpha[_ngcontent-%COMP%]{-webkit-box-flex:1;flex:1 1 0%;padding-left:6px}[dir=\"rtl\"][_nghost-%COMP%]   .sketch-single[_ngcontent-%COMP%], [dir=\"rtl\"]   [_nghost-%COMP%]   .sketch-single[_ngcontent-%COMP%]{padding-right:6px;padding-left:0}[dir=\"rtl\"][_nghost-%COMP%]   .sketch-alpha[_ngcontent-%COMP%], [dir=\"rtl\"]   [_nghost-%COMP%]   .sketch-alpha[_ngcontent-%COMP%]{padding-right:6px;padding-left:0}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SketchFieldsComponent, [{
		type: Component,
		args: [{
			selector: "color-sketch-fields",
			template: `
    <div class="sketch-fields">
      <div class="sketch-double">
        <color-editable-input
          [style]="{ input: input, label: label }"
          label="hex"
          [value]="hex.replace('#', '')"
          (onChange)="handleChange($event)"
        ></color-editable-input>
      </div>
      <div class="sketch-single">
        <color-editable-input
          [style]="{ input: input, label: label }"
          label="r"
          [value]="rgb.r"
          (onChange)="handleChange($event)"
          [dragLabel]="true"
          [dragMax]="255"
        ></color-editable-input>
      </div>
      <div class="sketch-single">
        <color-editable-input
          [style]="{ input: input, label: label }"
          label="g"
          [value]="rgb.g"
          (onChange)="handleChange($event)"
          [dragLabel]="true"
          [dragMax]="255"
        ></color-editable-input>
      </div>
      <div class="sketch-single">
        <color-editable-input
          [style]="{ input: input, label: label }"
          label="b"
          [value]="rgb.b"
          (onChange)="handleChange($event)"
          [dragLabel]="true"
          [dragMax]="255"
        ></color-editable-input>
      </div>
      @if (disableAlpha === false) {
        <div class="sketch-alpha">
          <color-editable-input
            [style]="{ input: input, label: label }"
            label="a"
            [value]="round(rgb.a * 100)"
            (onChange)="handleChange($event)"
            [dragLabel]="true"
            [dragMax]="100"
          ></color-editable-input>
        </div>
      }
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			preserveWhitespaces: false,
			standalone: false,
			styles: [".sketch-fields{display:flex;padding-top:4px}.sketch-double{-webkit-box-flex:2;flex:2 1 0%}.sketch-single{flex:1 1 0%;padding-left:6px}.sketch-alpha{-webkit-box-flex:1;flex:1 1 0%;padding-left:6px}:host-context([dir=\"rtl\"]) .sketch-single{padding-right:6px;padding-left:0}:host-context([dir=\"rtl\"]) .sketch-alpha{padding-right:6px;padding-left:0}\n"]
		}]
	}], null, {
		hsl: [{ type: Input }],
		rgb: [{ type: Input }],
		hex: [{ type: Input }],
		disableAlpha: [{ type: Input }],
		onChange: [{ type: Output }]
	});
})();
var SketchPresetColorsComponent = class SketchPresetColorsComponent {
	colors;
	onClick = new EventEmitter();
	onSwatchHover = new EventEmitter();
	swatchStyle = {
		borderRadius: "3px",
		boxShadow: "inset 0 0 0 1px rgba(0,0,0,.15)"
	};
	handleClick({ hex, $event }) {
		this.onClick.emit({
			hex,
			$event
		});
	}
	normalizeValue(val) {
		if (typeof val === "string") return { color: val };
		return val;
	}
	focusStyle(val) {
		return { boxShadow: `inset 0 0 0 1px rgba(0,0,0,.15), 0 0 4px ${this.normalizeValue(val).color}` };
	}
	static ɵfac = function SketchPresetColorsComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SketchPresetColorsComponent)();
	};
	static ɵcmp = (function() {
		function SketchPresetColorsComponent_For_2_Template(rf, ctx) {
			if (rf & 1) {
				const _r1 = ɵɵgetCurrentView();
				ɵɵelementStart(0, "div", 1)(1, "color-swatch", 2);
				ɵɵlistener("onClick", function SketchPresetColorsComponent_For_2_Template_color_swatch_onClick_1_listener($event) {
					ɵɵrestoreView(_r1);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.handleClick($event));
				})("onHover", function SketchPresetColorsComponent_For_2_Template_color_swatch_onHover_1_listener($event) {
					ɵɵrestoreView(_r1);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.onSwatchHover.emit($event));
				});
				ɵɵelementEnd()();
			}
			if (rf & 2) {
				const c_r3 = ctx.$implicit;
				const ctx_r1 = ɵɵnextContext();
				ɵɵadvance();
				ɵɵstyleMap(ctx_r1.swatchStyle);
				ɵɵproperty("color", ctx_r1.normalizeValue(c_r3).color)("focusStyle", ctx_r1.focusStyle(c_r3));
			}
		}
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: SketchPresetColorsComponent,
			selectors: [["color-sketch-preset-colors"]],
			inputs: { colors: "colors" },
			outputs: {
				onClick: "onClick",
				onSwatchHover: "onSwatchHover"
			},
			standalone: false,
			decls: 3,
			vars: 0,
			consts: [
				[1, "sketch-swatches"],
				[1, "sketch-wrap"],
				[
					1,
					"swatch",
					3,
					"onClick",
					"onHover",
					"color",
					"focusStyle"
				]
			],
			template: function SketchPresetColorsComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵelementStart(0, "div", 0);
					ɵɵrepeaterCreate(1, SketchPresetColorsComponent_For_2_Template, 2, 4, "div", 1, ɵɵrepeaterTrackByIdentity);
					ɵɵelementEnd();
				}
				if (rf & 2) {
					ɵɵadvance();
					ɵɵrepeater(ctx.colors);
				}
			},
			dependencies: [SwatchComponent],
			styles: [".sketch-swatches[_ngcontent-%COMP%]{position:relative;display:flex;flex-wrap:wrap;margin:0 -10px;padding:10px 0 0 10px;border-top:1px solid rgb(238,238,238)}.sketch-wrap[_ngcontent-%COMP%]{width:16px;height:16px;margin:0 10px 10px 0}[dir=\"rtl\"][_nghost-%COMP%]   .sketch-swatches[_ngcontent-%COMP%], [dir=\"rtl\"]   [_nghost-%COMP%]   .sketch-swatches[_ngcontent-%COMP%]{padding-right:10px;padding-left:0}[dir=\"rtl\"][_nghost-%COMP%]   .sketch-wrap[_ngcontent-%COMP%], [dir=\"rtl\"]   [_nghost-%COMP%]   .sketch-wrap[_ngcontent-%COMP%]{margin-left:10px;margin-right:0}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SketchPresetColorsComponent, [{
		type: Component,
		args: [{
			selector: "color-sketch-preset-colors",
			template: `
    <div class="sketch-swatches">
      @for (c of colors; track c) {
        <div class="sketch-wrap">
          <color-swatch
            [color]="normalizeValue(c).color"
            [style]="swatchStyle"
            [focusStyle]="focusStyle(c)"
            (onClick)="handleClick($event)"
            (onHover)="onSwatchHover.emit($event)"
            class="swatch"
          ></color-swatch>
        </div>
      }
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			preserveWhitespaces: false,
			standalone: false,
			styles: [".sketch-swatches{position:relative;display:flex;flex-wrap:wrap;margin:0 -10px;padding:10px 0 0 10px;border-top:1px solid rgb(238,238,238)}.sketch-wrap{width:16px;height:16px;margin:0 10px 10px 0}:host-context([dir=\"rtl\"]) .sketch-swatches{padding-right:10px;padding-left:0}:host-context([dir=\"rtl\"]) .sketch-wrap{margin-left:10px;margin-right:0}\n"]
		}]
	}], null, {
		colors: [{ type: Input }],
		onClick: [{ type: Output }],
		onSwatchHover: [{ type: Output }]
	});
})();
var SketchComponent = class SketchComponent extends ColorWrap {
	/** Remove alpha slider and options from picker */
	disableAlpha = false;
	/** Hex strings for default colors at bottom of picker */
	presetColors = [
		"#D0021B",
		"#F5A623",
		"#F8E71C",
		"#8B572A",
		"#7ED321",
		"#417505",
		"#BD10E0",
		"#9013FE",
		"#4A90E2",
		"#50E3C2",
		"#B8E986",
		"#000000",
		"#4A4A4A",
		"#9B9B9B",
		"#FFFFFF"
	];
	/** Width of picker */
	width = 200;
	activeBackground;
	constructor() {
		super();
	}
	afterValidChange() {
		const alpha = this.disableAlpha ? 1 : this.rgb.a;
		this.activeBackground = `rgba(${this.rgb.r}, ${this.rgb.g}, ${this.rgb.b}, ${alpha})`;
	}
	handleValueChange({ data, $event }) {
		this.handleChange(data, $event);
	}
	handleBlockChange({ hex, $event }) {
		if (isValidHex(hex)) this.handleChange({
			hex,
			source: "hex"
		}, $event);
	}
	static ɵfac = function SketchComponent_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || SketchComponent)();
	};
	static ɵcmp = (function() {
		function SketchComponent_Conditional_7_Template(rf, ctx) {
			if (rf & 1) {
				const _r1 = ɵɵgetCurrentView();
				ɵɵelementStart(0, "div", 6)(1, "color-alpha", 12);
				ɵɵlistener("onChange", function SketchComponent_Conditional_7_Template_color_alpha_onChange_1_listener($event) {
					ɵɵrestoreView(_r1);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.handleValueChange($event));
				});
				ɵɵelementEnd()();
			}
			if (rf & 2) {
				const ctx_r1 = ɵɵnextContext();
				ɵɵadvance();
				ɵɵproperty("radius", 2)("rgb", ctx_r1.rgb)("hsl", ctx_r1.hsl);
			}
		}
		function SketchComponent_Conditional_13_Template(rf, ctx) {
			if (rf & 1) {
				const _r3 = ɵɵgetCurrentView();
				ɵɵelementStart(0, "div", 11)(1, "color-sketch-preset-colors", 13);
				ɵɵlistener("onClick", function SketchComponent_Conditional_13_Template_color_sketch_preset_colors_onClick_1_listener($event) {
					ɵɵrestoreView(_r3);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.handleBlockChange($event));
				})("onSwatchHover", function SketchComponent_Conditional_13_Template_color_sketch_preset_colors_onSwatchHover_1_listener($event) {
					ɵɵrestoreView(_r3);
					const ctx_r1 = ɵɵnextContext();
					return ɵɵresetView(ctx_r1.onSwatchHover.emit($event));
				});
				ɵɵelementEnd()();
			}
			if (rf & 2) {
				const ctx_r1 = ɵɵnextContext();
				ɵɵadvance();
				ɵɵproperty("colors", ctx_r1.presetColors);
			}
		}
		return /*@__PURE__*/ ɵɵdefineComponent({
			type: SketchComponent,
			selectors: [["color-sketch"]],
			inputs: {
				disableAlpha: "disableAlpha",
				presetColors: "presetColors",
				width: "width"
			},
			standalone: false,
			features: [ɵɵProvidersFeature([{
				provide: NG_VALUE_ACCESSOR,
				useExisting: forwardRef(() => SketchComponent),
				multi: true
			}, {
				provide: ColorWrap,
				useExisting: forwardRef(() => SketchComponent)
			}]), ɵɵInheritDefinitionFeature],
			decls: 14,
			vars: 16,
			consts: [
				[1, "sketch-saturation"],
				[
					3,
					"onChange",
					"hsl",
					"hsv"
				],
				[1, "sketch-controls"],
				[1, "sketch-sliders"],
				[1, "sketch-hue"],
				[
					3,
					"onChange",
					"hsl"
				],
				[1, "sketch-alpha"],
				[1, "sketch-color"],
				[1, "sketch-active"],
				[1, "sketch-fields-container"],
				[
					3,
					"onChange",
					"rgb",
					"hsl",
					"hex",
					"disableAlpha"
				],
				[1, "sketch-swatches-container"],
				[
					3,
					"onChange",
					"radius",
					"rgb",
					"hsl"
				],
				[
					3,
					"onClick",
					"onSwatchHover",
					"colors"
				]
			],
			template: function SketchComponent_Template(rf, ctx) {
				if (rf & 1) {
					ɵɵelementStart(0, "div")(1, "div", 0)(2, "color-saturation", 1);
					ɵɵlistener("onChange", function SketchComponent_Template_color_saturation_onChange_2_listener($event) {
						return ctx.handleValueChange($event);
					});
					ɵɵelementEnd()();
					ɵɵelementStart(3, "div", 2)(4, "div", 3)(5, "div", 4)(6, "color-hue", 5);
					ɵɵlistener("onChange", function SketchComponent_Template_color_hue_onChange_6_listener($event) {
						return ctx.handleValueChange($event);
					});
					ɵɵelementEnd()();
					ɵɵconditionalCreate(7, SketchComponent_Conditional_7_Template, 2, 3, "div", 6);
					ɵɵelementEnd();
					ɵɵelementStart(8, "div", 7);
					ɵɵelement(9, "color-checkboard")(10, "div", 8);
					ɵɵelementEnd()();
					ɵɵelementStart(11, "div", 9)(12, "color-sketch-fields", 10);
					ɵɵlistener("onChange", function SketchComponent_Template_color_sketch_fields_onChange_12_listener($event) {
						return ctx.handleValueChange($event);
					});
					ɵɵelementEnd()();
					ɵɵconditionalCreate(13, SketchComponent_Conditional_13_Template, 2, 1, "div", 11);
					ɵɵelementEnd();
				}
				if (rf & 2) {
					ɵɵclassMap(ɵɵinterpolate1("sketch-picker ", ctx.className));
					ɵɵstyleProp("width", ctx.width, "px");
					ɵɵadvance(2);
					ɵɵproperty("hsl", ctx.hsl)("hsv", ctx.hsv);
					ɵɵadvance(4);
					ɵɵproperty("hsl", ctx.hsl);
					ɵɵadvance();
					ɵɵconditional(ctx.disableAlpha === false ? 7 : -1);
					ɵɵadvance(3);
					ɵɵstyleProp("background", ctx.activeBackground);
					ɵɵadvance(2);
					ɵɵproperty("rgb", ctx.rgb)("hsl", ctx.hsl)("hex", ctx.hex)("disableAlpha", ctx.disableAlpha);
					ɵɵadvance();
					ɵɵconditional(ctx.presetColors && ctx.presetColors.length ? 13 : -1);
				}
			},
			dependencies: () => [
				AlphaComponent,
				CheckboardComponent,
				HueComponent,
				SaturationComponent,
				SketchFieldsComponent,
				SketchPresetColorsComponent
			],
			styles: [".sketch-picker[_ngcontent-%COMP%]{padding:10px 10px 3px;box-sizing:initial;background:#fff;border-radius:4px;box-shadow:0 0 0 1px #00000026,0 8px 16px #00000026}.sketch-saturation[_ngcontent-%COMP%]{width:100%;padding-bottom:75%;position:relative;overflow:hidden}.sketch-fields-container[_ngcontent-%COMP%], .sketch-swatches-container[_ngcontent-%COMP%]{display:block}.sketch-controls[_ngcontent-%COMP%]{display:flex}.sketch-sliders[_ngcontent-%COMP%]{padding:4px 0;-webkit-box-flex:1;flex:1 1 0%}.sketch-hue[_ngcontent-%COMP%]{position:relative;height:10px;overflow:hidden}.sketch-alpha[_ngcontent-%COMP%]{position:relative;height:10px;margin-top:4px;overflow:hidden}.sketch-color[_ngcontent-%COMP%]{width:24px;height:24px;position:relative;margin-top:4px;margin-left:4px;border-radius:3px}.sketch-active[_ngcontent-%COMP%]{position:absolute;inset:0;border-radius:2px;box-shadow:#00000026 0 0 0 1px inset,#00000040 0 0 4px inset}[dir=\"rtl\"][_nghost-%COMP%]   .sketch-color[_ngcontent-%COMP%], [dir=\"rtl\"]   [_nghost-%COMP%]   .sketch-color[_ngcontent-%COMP%]{margin-right:4px;margin-left:0}"]
		});
	})();
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(SketchComponent, [{
		type: Component,
		args: [{
			selector: "color-sketch",
			template: `
    <div class="sketch-picker {{ className }}" [style.width.px]="width">
      <div class="sketch-saturation">
        <color-saturation [hsl]="hsl" [hsv]="hsv" (onChange)="handleValueChange($event)">
        </color-saturation>
      </div>
      <div class="sketch-controls">
        <div class="sketch-sliders">
          <div class="sketch-hue">
            <color-hue [hsl]="hsl" (onChange)="handleValueChange($event)"></color-hue>
          </div>
          @if (disableAlpha === false) {
            <div class="sketch-alpha">
              <color-alpha
                [radius]="2"
                [rgb]="rgb"
                [hsl]="hsl"
                (onChange)="handleValueChange($event)"
              ></color-alpha>
            </div>
          }
        </div>
        <div class="sketch-color">
          <color-checkboard></color-checkboard>
          <div class="sketch-active" [style.background]="activeBackground"></div>
        </div>
      </div>
      <div class="sketch-fields-container">
        <color-sketch-fields
          [rgb]="rgb"
          [hsl]="hsl"
          [hex]="hex"
          [disableAlpha]="disableAlpha"
          (onChange)="handleValueChange($event)"
        ></color-sketch-fields>
      </div>
      @if (presetColors && presetColors.length) {
        <div class="sketch-swatches-container">
          <color-sketch-preset-colors
            [colors]="presetColors"
            (onClick)="handleBlockChange($event)"
            (onSwatchHover)="onSwatchHover.emit($event)"
          ></color-sketch-preset-colors>
        </div>
      }
    </div>
  `,
			changeDetection: ChangeDetectionStrategy.OnPush,
			preserveWhitespaces: false,
			providers: [{
				provide: NG_VALUE_ACCESSOR,
				useExisting: forwardRef(() => SketchComponent),
				multi: true
			}, {
				provide: ColorWrap,
				useExisting: forwardRef(() => SketchComponent)
			}],
			standalone: false,
			styles: [".sketch-picker{padding:10px 10px 3px;box-sizing:initial;background:#fff;border-radius:4px;box-shadow:0 0 0 1px #00000026,0 8px 16px #00000026}.sketch-saturation{width:100%;padding-bottom:75%;position:relative;overflow:hidden}.sketch-fields-container,.sketch-swatches-container{display:block}.sketch-controls{display:flex}.sketch-sliders{padding:4px 0;-webkit-box-flex:1;flex:1 1 0%}.sketch-hue{position:relative;height:10px;overflow:hidden}.sketch-alpha{position:relative;height:10px;margin-top:4px;overflow:hidden}.sketch-color{width:24px;height:24px;position:relative;margin-top:4px;margin-left:4px;border-radius:3px}.sketch-active{position:absolute;inset:0;border-radius:2px;box-shadow:#00000026 0 0 0 1px inset,#00000040 0 0 4px inset}:host-context([dir=\"rtl\"]) .sketch-color{margin-right:4px;margin-left:0}\n"]
		}]
	}], () => [], {
		disableAlpha: [{ type: Input }],
		presetColors: [{ type: Input }],
		width: [{ type: Input }]
	});
})();
var ColorSketchModule = class ColorSketchModule {
	static ɵfac = function ColorSketchModule_Factory(__ngFactoryType__) {
		return new (__ngFactoryType__ || ColorSketchModule)();
	};
	static ɵmod = /*@__PURE__*/ ɵɵdefineNgModule({
		type: ColorSketchModule,
		declarations: [
			SketchComponent,
			SketchFieldsComponent,
			SketchPresetColorsComponent
		],
		imports: [
			CommonModule,
			AlphaModule,
			CheckboardModule,
			EditableInputModule,
			HueModule,
			SaturationModule,
			SwatchModule
		],
		exports: [
			SketchComponent,
			SketchFieldsComponent,
			SketchPresetColorsComponent
		]
	});
	static ɵinj = /*@__PURE__*/ ɵɵdefineInjector({ imports: [
		CommonModule,
		AlphaModule,
		CheckboardModule,
		EditableInputModule,
		HueModule,
		SaturationModule,
		SwatchModule
	] });
};
(() => {
	(typeof ngDevMode === "undefined" || ngDevMode) && setClassMetadata(ColorSketchModule, [{
		type: NgModule,
		args: [{
			declarations: [
				SketchComponent,
				SketchFieldsComponent,
				SketchPresetColorsComponent
			],
			exports: [
				SketchComponent,
				SketchFieldsComponent,
				SketchPresetColorsComponent
			],
			imports: [
				CommonModule,
				AlphaModule,
				CheckboardModule,
				EditableInputModule,
				HueModule,
				SaturationModule,
				SwatchModule
			]
		}]
	}], null, null);
})();
//#endregion
export { ColorSketchModule, SketchComponent, SketchFieldsComponent, SketchPresetColorsComponent };
