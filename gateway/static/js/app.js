(function(scope){
'use strict';

function F(arity, fun, wrapper) {
  wrapper.a = arity;
  wrapper.f = fun;
  return wrapper;
}

function F2(fun) {
  return F(2, fun, function(a) { return function(b) { return fun(a,b); }; })
}
function F3(fun) {
  return F(3, fun, function(a) {
    return function(b) { return function(c) { return fun(a, b, c); }; };
  });
}
function F4(fun) {
  return F(4, fun, function(a) { return function(b) { return function(c) {
    return function(d) { return fun(a, b, c, d); }; }; };
  });
}
function F5(fun) {
  return F(5, fun, function(a) { return function(b) { return function(c) {
    return function(d) { return function(e) { return fun(a, b, c, d, e); }; }; }; };
  });
}
function F6(fun) {
  return F(6, fun, function(a) { return function(b) { return function(c) {
    return function(d) { return function(e) { return function(f) {
    return fun(a, b, c, d, e, f); }; }; }; }; };
  });
}
function F7(fun) {
  return F(7, fun, function(a) { return function(b) { return function(c) {
    return function(d) { return function(e) { return function(f) {
    return function(g) { return fun(a, b, c, d, e, f, g); }; }; }; }; }; };
  });
}
function F8(fun) {
  return F(8, fun, function(a) { return function(b) { return function(c) {
    return function(d) { return function(e) { return function(f) {
    return function(g) { return function(h) {
    return fun(a, b, c, d, e, f, g, h); }; }; }; }; }; }; };
  });
}
function F9(fun) {
  return F(9, fun, function(a) { return function(b) { return function(c) {
    return function(d) { return function(e) { return function(f) {
    return function(g) { return function(h) { return function(i) {
    return fun(a, b, c, d, e, f, g, h, i); }; }; }; }; }; }; }; };
  });
}

function A2(fun, a, b) {
  return fun.a === 2 ? fun.f(a, b) : fun(a)(b);
}
function A3(fun, a, b, c) {
  return fun.a === 3 ? fun.f(a, b, c) : fun(a)(b)(c);
}
function A4(fun, a, b, c, d) {
  return fun.a === 4 ? fun.f(a, b, c, d) : fun(a)(b)(c)(d);
}
function A5(fun, a, b, c, d, e) {
  return fun.a === 5 ? fun.f(a, b, c, d, e) : fun(a)(b)(c)(d)(e);
}
function A6(fun, a, b, c, d, e, f) {
  return fun.a === 6 ? fun.f(a, b, c, d, e, f) : fun(a)(b)(c)(d)(e)(f);
}
function A7(fun, a, b, c, d, e, f, g) {
  return fun.a === 7 ? fun.f(a, b, c, d, e, f, g) : fun(a)(b)(c)(d)(e)(f)(g);
}
function A8(fun, a, b, c, d, e, f, g, h) {
  return fun.a === 8 ? fun.f(a, b, c, d, e, f, g, h) : fun(a)(b)(c)(d)(e)(f)(g)(h);
}
function A9(fun, a, b, c, d, e, f, g, h, i) {
  return fun.a === 9 ? fun.f(a, b, c, d, e, f, g, h, i) : fun(a)(b)(c)(d)(e)(f)(g)(h)(i);
}

console.warn('Compiled in DEV mode. Follow the advice at https://elm-lang.org/0.19.0/optimize for better performance and smaller assets.');


var _List_Nil_UNUSED = { $: 0 };
var _List_Nil = { $: '[]' };

function _List_Cons_UNUSED(hd, tl) { return { $: 1, a: hd, b: tl }; }
function _List_Cons(hd, tl) { return { $: '::', a: hd, b: tl }; }


var _List_cons = F2(_List_Cons);

function _List_fromArray(arr)
{
	var out = _List_Nil;
	for (var i = arr.length; i--; )
	{
		out = _List_Cons(arr[i], out);
	}
	return out;
}

function _List_toArray(xs)
{
	for (var out = []; xs.b; xs = xs.b) // WHILE_CONS
	{
		out.push(xs.a);
	}
	return out;
}

var _List_map2 = F3(function(f, xs, ys)
{
	for (var arr = []; xs.b && ys.b; xs = xs.b, ys = ys.b) // WHILE_CONSES
	{
		arr.push(A2(f, xs.a, ys.a));
	}
	return _List_fromArray(arr);
});

var _List_map3 = F4(function(f, xs, ys, zs)
{
	for (var arr = []; xs.b && ys.b && zs.b; xs = xs.b, ys = ys.b, zs = zs.b) // WHILE_CONSES
	{
		arr.push(A3(f, xs.a, ys.a, zs.a));
	}
	return _List_fromArray(arr);
});

var _List_map4 = F5(function(f, ws, xs, ys, zs)
{
	for (var arr = []; ws.b && xs.b && ys.b && zs.b; ws = ws.b, xs = xs.b, ys = ys.b, zs = zs.b) // WHILE_CONSES
	{
		arr.push(A4(f, ws.a, xs.a, ys.a, zs.a));
	}
	return _List_fromArray(arr);
});

var _List_map5 = F6(function(f, vs, ws, xs, ys, zs)
{
	for (var arr = []; vs.b && ws.b && xs.b && ys.b && zs.b; vs = vs.b, ws = ws.b, xs = xs.b, ys = ys.b, zs = zs.b) // WHILE_CONSES
	{
		arr.push(A5(f, vs.a, ws.a, xs.a, ys.a, zs.a));
	}
	return _List_fromArray(arr);
});

var _List_sortBy = F2(function(f, xs)
{
	return _List_fromArray(_List_toArray(xs).sort(function(a, b) {
		return _Utils_cmp(f(a), f(b));
	}));
});

var _List_sortWith = F2(function(f, xs)
{
	return _List_fromArray(_List_toArray(xs).sort(function(a, b) {
		var ord = A2(f, a, b);
		return ord === elm$core$Basics$EQ ? 0 : ord === elm$core$Basics$LT ? -1 : 1;
	}));
});



// EQUALITY

function _Utils_eq(x, y)
{
	for (
		var pair, stack = [], isEqual = _Utils_eqHelp(x, y, 0, stack);
		isEqual && (pair = stack.pop());
		isEqual = _Utils_eqHelp(pair.a, pair.b, 0, stack)
		)
	{}

	return isEqual;
}

function _Utils_eqHelp(x, y, depth, stack)
{
	if (depth > 100)
	{
		stack.push(_Utils_Tuple2(x,y));
		return true;
	}

	if (x === y)
	{
		return true;
	}

	if (typeof x !== 'object' || x === null || y === null)
	{
		typeof x === 'function' && _Debug_crash(5);
		return false;
	}

	/**/
	if (x.$ === 'Set_elm_builtin')
	{
		x = elm$core$Set$toList(x);
		y = elm$core$Set$toList(y);
	}
	if (x.$ === 'RBNode_elm_builtin' || x.$ === 'RBEmpty_elm_builtin')
	{
		x = elm$core$Dict$toList(x);
		y = elm$core$Dict$toList(y);
	}
	//*/

	/**_UNUSED/
	if (x.$ < 0)
	{
		x = elm$core$Dict$toList(x);
		y = elm$core$Dict$toList(y);
	}
	//*/

	for (var key in x)
	{
		if (!_Utils_eqHelp(x[key], y[key], depth + 1, stack))
		{
			return false;
		}
	}
	return true;
}

var _Utils_equal = F2(_Utils_eq);
var _Utils_notEqual = F2(function(a, b) { return !_Utils_eq(a,b); });



// COMPARISONS

// Code in Generate/JavaScript.hs, Basics.js, and List.js depends on
// the particular integer values assigned to LT, EQ, and GT.

function _Utils_cmp(x, y, ord)
{
	if (typeof x !== 'object')
	{
		return x === y ? /*EQ*/ 0 : x < y ? /*LT*/ -1 : /*GT*/ 1;
	}

	/**/
	if (x instanceof String)
	{
		var a = x.valueOf();
		var b = y.valueOf();
		return a === b ? 0 : a < b ? -1 : 1;
	}
	//*/

	/**_UNUSED/
	if (typeof x.$ === 'undefined')
	//*/
	/**/
	if (x.$[0] === '#')
	//*/
	{
		return (ord = _Utils_cmp(x.a, y.a))
			? ord
			: (ord = _Utils_cmp(x.b, y.b))
				? ord
				: _Utils_cmp(x.c, y.c);
	}

	// traverse conses until end of a list or a mismatch
	for (; x.b && y.b && !(ord = _Utils_cmp(x.a, y.a)); x = x.b, y = y.b) {} // WHILE_CONSES
	return ord || (x.b ? /*GT*/ 1 : y.b ? /*LT*/ -1 : /*EQ*/ 0);
}

var _Utils_lt = F2(function(a, b) { return _Utils_cmp(a, b) < 0; });
var _Utils_le = F2(function(a, b) { return _Utils_cmp(a, b) < 1; });
var _Utils_gt = F2(function(a, b) { return _Utils_cmp(a, b) > 0; });
var _Utils_ge = F2(function(a, b) { return _Utils_cmp(a, b) >= 0; });

var _Utils_compare = F2(function(x, y)
{
	var n = _Utils_cmp(x, y);
	return n < 0 ? elm$core$Basics$LT : n ? elm$core$Basics$GT : elm$core$Basics$EQ;
});


// COMMON VALUES

var _Utils_Tuple0_UNUSED = 0;
var _Utils_Tuple0 = { $: '#0' };

function _Utils_Tuple2_UNUSED(a, b) { return { a: a, b: b }; }
function _Utils_Tuple2(a, b) { return { $: '#2', a: a, b: b }; }

function _Utils_Tuple3_UNUSED(a, b, c) { return { a: a, b: b, c: c }; }
function _Utils_Tuple3(a, b, c) { return { $: '#3', a: a, b: b, c: c }; }

function _Utils_chr_UNUSED(c) { return c; }
function _Utils_chr(c) { return new String(c); }


// RECORDS

function _Utils_update(oldRecord, updatedFields)
{
	var newRecord = {};

	for (var key in oldRecord)
	{
		newRecord[key] = oldRecord[key];
	}

	for (var key in updatedFields)
	{
		newRecord[key] = updatedFields[key];
	}

	return newRecord;
}


// APPEND

var _Utils_append = F2(_Utils_ap);

function _Utils_ap(xs, ys)
{
	// append Strings
	if (typeof xs === 'string')
	{
		return xs + ys;
	}

	// append Lists
	if (!xs.b)
	{
		return ys;
	}
	var root = _List_Cons(xs.a, ys);
	xs = xs.b
	for (var curr = root; xs.b; xs = xs.b) // WHILE_CONS
	{
		curr = curr.b = _List_Cons(xs.a, ys);
	}
	return root;
}



var _JsArray_empty = [];

function _JsArray_singleton(value)
{
    return [value];
}

function _JsArray_length(array)
{
    return array.length;
}

var _JsArray_initialize = F3(function(size, offset, func)
{
    var result = new Array(size);

    for (var i = 0; i < size; i++)
    {
        result[i] = func(offset + i);
    }

    return result;
});

var _JsArray_initializeFromList = F2(function (max, ls)
{
    var result = new Array(max);

    for (var i = 0; i < max && ls.b; i++)
    {
        result[i] = ls.a;
        ls = ls.b;
    }

    result.length = i;
    return _Utils_Tuple2(result, ls);
});

var _JsArray_unsafeGet = F2(function(index, array)
{
    return array[index];
});

var _JsArray_unsafeSet = F3(function(index, value, array)
{
    var length = array.length;
    var result = new Array(length);

    for (var i = 0; i < length; i++)
    {
        result[i] = array[i];
    }

    result[index] = value;
    return result;
});

var _JsArray_push = F2(function(value, array)
{
    var length = array.length;
    var result = new Array(length + 1);

    for (var i = 0; i < length; i++)
    {
        result[i] = array[i];
    }

    result[length] = value;
    return result;
});

var _JsArray_foldl = F3(function(func, acc, array)
{
    var length = array.length;

    for (var i = 0; i < length; i++)
    {
        acc = A2(func, array[i], acc);
    }

    return acc;
});

var _JsArray_foldr = F3(function(func, acc, array)
{
    for (var i = array.length - 1; i >= 0; i--)
    {
        acc = A2(func, array[i], acc);
    }

    return acc;
});

var _JsArray_map = F2(function(func, array)
{
    var length = array.length;
    var result = new Array(length);

    for (var i = 0; i < length; i++)
    {
        result[i] = func(array[i]);
    }

    return result;
});

var _JsArray_indexedMap = F3(function(func, offset, array)
{
    var length = array.length;
    var result = new Array(length);

    for (var i = 0; i < length; i++)
    {
        result[i] = A2(func, offset + i, array[i]);
    }

    return result;
});

var _JsArray_slice = F3(function(from, to, array)
{
    return array.slice(from, to);
});

var _JsArray_appendN = F3(function(n, dest, source)
{
    var destLen = dest.length;
    var itemsToCopy = n - destLen;

    if (itemsToCopy > source.length)
    {
        itemsToCopy = source.length;
    }

    var size = destLen + itemsToCopy;
    var result = new Array(size);

    for (var i = 0; i < destLen; i++)
    {
        result[i] = dest[i];
    }

    for (var i = 0; i < itemsToCopy; i++)
    {
        result[i + destLen] = source[i];
    }

    return result;
});



// LOG

var _Debug_log_UNUSED = F2(function(tag, value)
{
	return value;
});

var _Debug_log = F2(function(tag, value)
{
	console.log(tag + ': ' + _Debug_toString(value));
	return value;
});


// TODOS

function _Debug_todo(moduleName, region)
{
	return function(message) {
		_Debug_crash(8, moduleName, region, message);
	};
}

function _Debug_todoCase(moduleName, region, value)
{
	return function(message) {
		_Debug_crash(9, moduleName, region, value, message);
	};
}


// TO STRING

function _Debug_toString_UNUSED(value)
{
	return '<internals>';
}

function _Debug_toString(value)
{
	return _Debug_toAnsiString(false, value);
}

function _Debug_toAnsiString(ansi, value)
{
	if (typeof value === 'function')
	{
		return _Debug_internalColor(ansi, '<function>');
	}

	if (typeof value === 'boolean')
	{
		return _Debug_ctorColor(ansi, value ? 'True' : 'False');
	}

	if (typeof value === 'number')
	{
		return _Debug_numberColor(ansi, value + '');
	}

	if (value instanceof String)
	{
		return _Debug_charColor(ansi, "'" + _Debug_addSlashes(value, true) + "'");
	}

	if (typeof value === 'string')
	{
		return _Debug_stringColor(ansi, '"' + _Debug_addSlashes(value, false) + '"');
	}

	if (typeof value === 'object' && '$' in value)
	{
		var tag = value.$;

		if (typeof tag === 'number')
		{
			return _Debug_internalColor(ansi, '<internals>');
		}

		if (tag[0] === '#')
		{
			var output = [];
			for (var k in value)
			{
				if (k === '$') continue;
				output.push(_Debug_toAnsiString(ansi, value[k]));
			}
			return '(' + output.join(',') + ')';
		}

		if (tag === 'Set_elm_builtin')
		{
			return _Debug_ctorColor(ansi, 'Set')
				+ _Debug_fadeColor(ansi, '.fromList') + ' '
				+ _Debug_toAnsiString(ansi, elm$core$Set$toList(value));
		}

		if (tag === 'RBNode_elm_builtin' || tag === 'RBEmpty_elm_builtin')
		{
			return _Debug_ctorColor(ansi, 'Dict')
				+ _Debug_fadeColor(ansi, '.fromList') + ' '
				+ _Debug_toAnsiString(ansi, elm$core$Dict$toList(value));
		}

		if (tag === 'Array_elm_builtin')
		{
			return _Debug_ctorColor(ansi, 'Array')
				+ _Debug_fadeColor(ansi, '.fromList') + ' '
				+ _Debug_toAnsiString(ansi, elm$core$Array$toList(value));
		}

		if (tag === '::' || tag === '[]')
		{
			var output = '[';

			value.b && (output += _Debug_toAnsiString(ansi, value.a), value = value.b)

			for (; value.b; value = value.b) // WHILE_CONS
			{
				output += ',' + _Debug_toAnsiString(ansi, value.a);
			}
			return output + ']';
		}

		var output = '';
		for (var i in value)
		{
			if (i === '$') continue;
			var str = _Debug_toAnsiString(ansi, value[i]);
			var c0 = str[0];
			var parenless = c0 === '{' || c0 === '(' || c0 === '[' || c0 === '<' || c0 === '"' || str.indexOf(' ') < 0;
			output += ' ' + (parenless ? str : '(' + str + ')');
		}
		return _Debug_ctorColor(ansi, tag) + output;
	}

	if (typeof DataView === 'function' && value instanceof DataView)
	{
		return _Debug_stringColor(ansi, '<' + value.byteLength + ' bytes>');
	}

	if (typeof File === 'function' && value instanceof File)
	{
		return _Debug_internalColor(ansi, '<' + value.name + '>');
	}

	if (typeof value === 'object')
	{
		var output = [];
		for (var key in value)
		{
			var field = key[0] === '_' ? key.slice(1) : key;
			output.push(_Debug_fadeColor(ansi, field) + ' = ' + _Debug_toAnsiString(ansi, value[key]));
		}
		if (output.length === 0)
		{
			return '{}';
		}
		return '{ ' + output.join(', ') + ' }';
	}

	return _Debug_internalColor(ansi, '<internals>');
}

function _Debug_addSlashes(str, isChar)
{
	var s = str
		.replace(/\\/g, '\\\\')
		.replace(/\n/g, '\\n')
		.replace(/\t/g, '\\t')
		.replace(/\r/g, '\\r')
		.replace(/\v/g, '\\v')
		.replace(/\0/g, '\\0');

	if (isChar)
	{
		return s.replace(/\'/g, '\\\'');
	}
	else
	{
		return s.replace(/\"/g, '\\"');
	}
}

function _Debug_ctorColor(ansi, string)
{
	return ansi ? '\x1b[96m' + string + '\x1b[0m' : string;
}

function _Debug_numberColor(ansi, string)
{
	return ansi ? '\x1b[95m' + string + '\x1b[0m' : string;
}

function _Debug_stringColor(ansi, string)
{
	return ansi ? '\x1b[93m' + string + '\x1b[0m' : string;
}

function _Debug_charColor(ansi, string)
{
	return ansi ? '\x1b[92m' + string + '\x1b[0m' : string;
}

function _Debug_fadeColor(ansi, string)
{
	return ansi ? '\x1b[37m' + string + '\x1b[0m' : string;
}

function _Debug_internalColor(ansi, string)
{
	return ansi ? '\x1b[94m' + string + '\x1b[0m' : string;
}

function _Debug_toHexDigit(n)
{
	return String.fromCharCode(n < 10 ? 48 + n : 55 + n);
}


// CRASH


function _Debug_crash_UNUSED(identifier)
{
	throw new Error('https://github.com/elm/core/blob/1.0.0/hints/' + identifier + '.md');
}


function _Debug_crash(identifier, fact1, fact2, fact3, fact4)
{
	switch(identifier)
	{
		case 0:
			throw new Error('What node should I take over? In JavaScript I need something like:\n\n    Elm.Main.init({\n        node: document.getElementById("elm-node")\n    })\n\nYou need to do this with any Browser.sandbox or Browser.element program.');

		case 1:
			throw new Error('Browser.application programs cannot handle URLs like this:\n\n    ' + document.location.href + '\n\nWhat is the root? The root of your file system? Try looking at this program with `elm reactor` or some other server.');

		case 2:
			var jsonErrorString = fact1;
			throw new Error('Problem with the flags given to your Elm program on initialization.\n\n' + jsonErrorString);

		case 3:
			var portName = fact1;
			throw new Error('There can only be one port named `' + portName + '`, but your program has multiple.');

		case 4:
			var portName = fact1;
			var problem = fact2;
			throw new Error('Trying to send an unexpected type of value through port `' + portName + '`:\n' + problem);

		case 5:
			throw new Error('Trying to use `(==)` on functions.\nThere is no way to know if functions are "the same" in the Elm sense.\nRead more about this at https://package.elm-lang.org/packages/elm/core/latest/Basics#== which describes why it is this way and what the better version will look like.');

		case 6:
			var moduleName = fact1;
			throw new Error('Your page is loading multiple Elm scripts with a module named ' + moduleName + '. Maybe a duplicate script is getting loaded accidentally? If not, rename one of them so I know which is which!');

		case 8:
			var moduleName = fact1;
			var region = fact2;
			var message = fact3;
			throw new Error('TODO in module `' + moduleName + '` ' + _Debug_regionToString(region) + '\n\n' + message);

		case 9:
			var moduleName = fact1;
			var region = fact2;
			var value = fact3;
			var message = fact4;
			throw new Error(
				'TODO in module `' + moduleName + '` from the `case` expression '
				+ _Debug_regionToString(region) + '\n\nIt received the following value:\n\n    '
				+ _Debug_toString(value).replace('\n', '\n    ')
				+ '\n\nBut the branch that handles it says:\n\n    ' + message.replace('\n', '\n    ')
			);

		case 10:
			throw new Error('Bug in https://github.com/elm/virtual-dom/issues');

		case 11:
			throw new Error('Cannot perform mod 0. Division by zero error.');
	}
}

function _Debug_regionToString(region)
{
	if (region.start.line === region.end.line)
	{
		return 'on line ' + region.start.line;
	}
	return 'on lines ' + region.start.line + ' through ' + region.end.line;
}



// MATH

var _Basics_add = F2(function(a, b) { return a + b; });
var _Basics_sub = F2(function(a, b) { return a - b; });
var _Basics_mul = F2(function(a, b) { return a * b; });
var _Basics_fdiv = F2(function(a, b) { return a / b; });
var _Basics_idiv = F2(function(a, b) { return (a / b) | 0; });
var _Basics_pow = F2(Math.pow);

var _Basics_remainderBy = F2(function(b, a) { return a % b; });

// https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/divmodnote-letter.pdf
var _Basics_modBy = F2(function(modulus, x)
{
	var answer = x % modulus;
	return modulus === 0
		? _Debug_crash(11)
		:
	((answer > 0 && modulus < 0) || (answer < 0 && modulus > 0))
		? answer + modulus
		: answer;
});


// TRIGONOMETRY

var _Basics_pi = Math.PI;
var _Basics_e = Math.E;
var _Basics_cos = Math.cos;
var _Basics_sin = Math.sin;
var _Basics_tan = Math.tan;
var _Basics_acos = Math.acos;
var _Basics_asin = Math.asin;
var _Basics_atan = Math.atan;
var _Basics_atan2 = F2(Math.atan2);


// MORE MATH

function _Basics_toFloat(x) { return x; }
function _Basics_truncate(n) { return n | 0; }
function _Basics_isInfinite(n) { return n === Infinity || n === -Infinity; }

var _Basics_ceiling = Math.ceil;
var _Basics_floor = Math.floor;
var _Basics_round = Math.round;
var _Basics_sqrt = Math.sqrt;
var _Basics_log = Math.log;
var _Basics_isNaN = isNaN;


// BOOLEANS

function _Basics_not(bool) { return !bool; }
var _Basics_and = F2(function(a, b) { return a && b; });
var _Basics_or  = F2(function(a, b) { return a || b; });
var _Basics_xor = F2(function(a, b) { return a !== b; });



function _Char_toCode(char)
{
	var code = char.charCodeAt(0);
	if (0xD800 <= code && code <= 0xDBFF)
	{
		return (code - 0xD800) * 0x400 + char.charCodeAt(1) - 0xDC00 + 0x10000
	}
	return code;
}

function _Char_fromCode(code)
{
	return _Utils_chr(
		(code < 0 || 0x10FFFF < code)
			? '\uFFFD'
			:
		(code <= 0xFFFF)
			? String.fromCharCode(code)
			:
		(code -= 0x10000,
			String.fromCharCode(Math.floor(code / 0x400) + 0xD800, code % 0x400 + 0xDC00)
		)
	);
}

function _Char_toUpper(char)
{
	return _Utils_chr(char.toUpperCase());
}

function _Char_toLower(char)
{
	return _Utils_chr(char.toLowerCase());
}

function _Char_toLocaleUpper(char)
{
	return _Utils_chr(char.toLocaleUpperCase());
}

function _Char_toLocaleLower(char)
{
	return _Utils_chr(char.toLocaleLowerCase());
}



var _String_cons = F2(function(chr, str)
{
	return chr + str;
});

function _String_uncons(string)
{
	var word = string.charCodeAt(0);
	return word
		? elm$core$Maybe$Just(
			0xD800 <= word && word <= 0xDBFF
				? _Utils_Tuple2(_Utils_chr(string[0] + string[1]), string.slice(2))
				: _Utils_Tuple2(_Utils_chr(string[0]), string.slice(1))
		)
		: elm$core$Maybe$Nothing;
}

var _String_append = F2(function(a, b)
{
	return a + b;
});

function _String_length(str)
{
	return str.length;
}

var _String_map = F2(function(func, string)
{
	var len = string.length;
	var array = new Array(len);
	var i = 0;
	while (i < len)
	{
		var word = string.charCodeAt(i);
		if (0xD800 <= word && word <= 0xDBFF)
		{
			array[i] = func(_Utils_chr(string[i] + string[i+1]));
			i += 2;
			continue;
		}
		array[i] = func(_Utils_chr(string[i]));
		i++;
	}
	return array.join('');
});

var _String_filter = F2(function(isGood, str)
{
	var arr = [];
	var len = str.length;
	var i = 0;
	while (i < len)
	{
		var char = str[i];
		var word = str.charCodeAt(i);
		i++;
		if (0xD800 <= word && word <= 0xDBFF)
		{
			char += str[i];
			i++;
		}

		if (isGood(_Utils_chr(char)))
		{
			arr.push(char);
		}
	}
	return arr.join('');
});

function _String_reverse(str)
{
	var len = str.length;
	var arr = new Array(len);
	var i = 0;
	while (i < len)
	{
		var word = str.charCodeAt(i);
		if (0xD800 <= word && word <= 0xDBFF)
		{
			arr[len - i] = str[i + 1];
			i++;
			arr[len - i] = str[i - 1];
			i++;
		}
		else
		{
			arr[len - i] = str[i];
			i++;
		}
	}
	return arr.join('');
}

var _String_foldl = F3(function(func, state, string)
{
	var len = string.length;
	var i = 0;
	while (i < len)
	{
		var char = string[i];
		var word = string.charCodeAt(i);
		i++;
		if (0xD800 <= word && word <= 0xDBFF)
		{
			char += string[i];
			i++;
		}
		state = A2(func, _Utils_chr(char), state);
	}
	return state;
});

var _String_foldr = F3(function(func, state, string)
{
	var i = string.length;
	while (i--)
	{
		var char = string[i];
		var word = string.charCodeAt(i);
		if (0xDC00 <= word && word <= 0xDFFF)
		{
			i--;
			char = string[i] + char;
		}
		state = A2(func, _Utils_chr(char), state);
	}
	return state;
});

var _String_split = F2(function(sep, str)
{
	return str.split(sep);
});

var _String_join = F2(function(sep, strs)
{
	return strs.join(sep);
});

var _String_slice = F3(function(start, end, str) {
	return str.slice(start, end);
});

function _String_trim(str)
{
	return str.trim();
}

function _String_trimLeft(str)
{
	return str.replace(/^\s+/, '');
}

function _String_trimRight(str)
{
	return str.replace(/\s+$/, '');
}

function _String_words(str)
{
	return _List_fromArray(str.trim().split(/\s+/g));
}

function _String_lines(str)
{
	return _List_fromArray(str.split(/\r\n|\r|\n/g));
}

function _String_toUpper(str)
{
	return str.toUpperCase();
}

function _String_toLower(str)
{
	return str.toLowerCase();
}

var _String_any = F2(function(isGood, string)
{
	var i = string.length;
	while (i--)
	{
		var char = string[i];
		var word = string.charCodeAt(i);
		if (0xDC00 <= word && word <= 0xDFFF)
		{
			i--;
			char = string[i] + char;
		}
		if (isGood(_Utils_chr(char)))
		{
			return true;
		}
	}
	return false;
});

var _String_all = F2(function(isGood, string)
{
	var i = string.length;
	while (i--)
	{
		var char = string[i];
		var word = string.charCodeAt(i);
		if (0xDC00 <= word && word <= 0xDFFF)
		{
			i--;
			char = string[i] + char;
		}
		if (!isGood(_Utils_chr(char)))
		{
			return false;
		}
	}
	return true;
});

var _String_contains = F2(function(sub, str)
{
	return str.indexOf(sub) > -1;
});

var _String_startsWith = F2(function(sub, str)
{
	return str.indexOf(sub) === 0;
});

var _String_endsWith = F2(function(sub, str)
{
	return str.length >= sub.length &&
		str.lastIndexOf(sub) === str.length - sub.length;
});

var _String_indexes = F2(function(sub, str)
{
	var subLen = sub.length;

	if (subLen < 1)
	{
		return _List_Nil;
	}

	var i = 0;
	var is = [];

	while ((i = str.indexOf(sub, i)) > -1)
	{
		is.push(i);
		i = i + subLen;
	}

	return _List_fromArray(is);
});


// TO STRING

function _String_fromNumber(number)
{
	return number + '';
}


// INT CONVERSIONS

function _String_toInt(str)
{
	var total = 0;
	var code0 = str.charCodeAt(0);
	var start = code0 == 0x2B /* + */ || code0 == 0x2D /* - */ ? 1 : 0;

	for (var i = start; i < str.length; ++i)
	{
		var code = str.charCodeAt(i);
		if (code < 0x30 || 0x39 < code)
		{
			return elm$core$Maybe$Nothing;
		}
		total = 10 * total + code - 0x30;
	}

	return i == start
		? elm$core$Maybe$Nothing
		: elm$core$Maybe$Just(code0 == 0x2D ? -total : total);
}


// FLOAT CONVERSIONS

function _String_toFloat(s)
{
	// check if it is a hex, octal, or binary number
	if (s.length === 0 || /[\sxbo]/.test(s))
	{
		return elm$core$Maybe$Nothing;
	}
	var n = +s;
	// faster isNaN check
	return n === n ? elm$core$Maybe$Just(n) : elm$core$Maybe$Nothing;
}

function _String_fromList(chars)
{
	return _List_toArray(chars).join('');
}




/**/
function _Json_errorToString(error)
{
	return elm$json$Json$Decode$errorToString(error);
}
//*/


// CORE DECODERS

function _Json_succeed(msg)
{
	return {
		$: 0,
		a: msg
	};
}

function _Json_fail(msg)
{
	return {
		$: 1,
		a: msg
	};
}

function _Json_decodePrim(decoder)
{
	return { $: 2, b: decoder };
}

var _Json_decodeInt = _Json_decodePrim(function(value) {
	return (typeof value !== 'number')
		? _Json_expecting('an INT', value)
		:
	(-2147483647 < value && value < 2147483647 && (value | 0) === value)
		? elm$core$Result$Ok(value)
		:
	(isFinite(value) && !(value % 1))
		? elm$core$Result$Ok(value)
		: _Json_expecting('an INT', value);
});

var _Json_decodeBool = _Json_decodePrim(function(value) {
	return (typeof value === 'boolean')
		? elm$core$Result$Ok(value)
		: _Json_expecting('a BOOL', value);
});

var _Json_decodeFloat = _Json_decodePrim(function(value) {
	return (typeof value === 'number')
		? elm$core$Result$Ok(value)
		: _Json_expecting('a FLOAT', value);
});

var _Json_decodeValue = _Json_decodePrim(function(value) {
	return elm$core$Result$Ok(_Json_wrap(value));
});

var _Json_decodeString = _Json_decodePrim(function(value) {
	return (typeof value === 'string')
		? elm$core$Result$Ok(value)
		: (value instanceof String)
			? elm$core$Result$Ok(value + '')
			: _Json_expecting('a STRING', value);
});

function _Json_decodeList(decoder) { return { $: 3, b: decoder }; }
function _Json_decodeArray(decoder) { return { $: 4, b: decoder }; }

function _Json_decodeNull(value) { return { $: 5, c: value }; }

var _Json_decodeField = F2(function(field, decoder)
{
	return {
		$: 6,
		d: field,
		b: decoder
	};
});

var _Json_decodeIndex = F2(function(index, decoder)
{
	return {
		$: 7,
		e: index,
		b: decoder
	};
});

function _Json_decodeKeyValuePairs(decoder)
{
	return {
		$: 8,
		b: decoder
	};
}

function _Json_mapMany(f, decoders)
{
	return {
		$: 9,
		f: f,
		g: decoders
	};
}

var _Json_andThen = F2(function(callback, decoder)
{
	return {
		$: 10,
		b: decoder,
		h: callback
	};
});

function _Json_oneOf(decoders)
{
	return {
		$: 11,
		g: decoders
	};
}


// DECODING OBJECTS

var _Json_map1 = F2(function(f, d1)
{
	return _Json_mapMany(f, [d1]);
});

var _Json_map2 = F3(function(f, d1, d2)
{
	return _Json_mapMany(f, [d1, d2]);
});

var _Json_map3 = F4(function(f, d1, d2, d3)
{
	return _Json_mapMany(f, [d1, d2, d3]);
});

var _Json_map4 = F5(function(f, d1, d2, d3, d4)
{
	return _Json_mapMany(f, [d1, d2, d3, d4]);
});

var _Json_map5 = F6(function(f, d1, d2, d3, d4, d5)
{
	return _Json_mapMany(f, [d1, d2, d3, d4, d5]);
});

var _Json_map6 = F7(function(f, d1, d2, d3, d4, d5, d6)
{
	return _Json_mapMany(f, [d1, d2, d3, d4, d5, d6]);
});

var _Json_map7 = F8(function(f, d1, d2, d3, d4, d5, d6, d7)
{
	return _Json_mapMany(f, [d1, d2, d3, d4, d5, d6, d7]);
});

var _Json_map8 = F9(function(f, d1, d2, d3, d4, d5, d6, d7, d8)
{
	return _Json_mapMany(f, [d1, d2, d3, d4, d5, d6, d7, d8]);
});


// DECODE

var _Json_runOnString = F2(function(decoder, string)
{
	try
	{
		var value = JSON.parse(string);
		return _Json_runHelp(decoder, value);
	}
	catch (e)
	{
		return elm$core$Result$Err(A2(elm$json$Json$Decode$Failure, 'This is not valid JSON! ' + e.message, _Json_wrap(string)));
	}
});

var _Json_run = F2(function(decoder, value)
{
	return _Json_runHelp(decoder, _Json_unwrap(value));
});

function _Json_runHelp(decoder, value)
{
	switch (decoder.$)
	{
		case 2:
			return decoder.b(value);

		case 5:
			return (value === null)
				? elm$core$Result$Ok(decoder.c)
				: _Json_expecting('null', value);

		case 3:
			if (!_Json_isArray(value))
			{
				return _Json_expecting('a LIST', value);
			}
			return _Json_runArrayDecoder(decoder.b, value, _List_fromArray);

		case 4:
			if (!_Json_isArray(value))
			{
				return _Json_expecting('an ARRAY', value);
			}
			return _Json_runArrayDecoder(decoder.b, value, _Json_toElmArray);

		case 6:
			var field = decoder.d;
			if (typeof value !== 'object' || value === null || !(field in value))
			{
				return _Json_expecting('an OBJECT with a field named `' + field + '`', value);
			}
			var result = _Json_runHelp(decoder.b, value[field]);
			return (elm$core$Result$isOk(result)) ? result : elm$core$Result$Err(A2(elm$json$Json$Decode$Field, field, result.a));

		case 7:
			var index = decoder.e;
			if (!_Json_isArray(value))
			{
				return _Json_expecting('an ARRAY', value);
			}
			if (index >= value.length)
			{
				return _Json_expecting('a LONGER array. Need index ' + index + ' but only see ' + value.length + ' entries', value);
			}
			var result = _Json_runHelp(decoder.b, value[index]);
			return (elm$core$Result$isOk(result)) ? result : elm$core$Result$Err(A2(elm$json$Json$Decode$Index, index, result.a));

		case 8:
			if (typeof value !== 'object' || value === null || !_Json_isArray(value))
			{
				return _Json_expecting('an OBJECT', value);
			}

			var keyValuePairs = _List_Nil;
			// TODO test perf of Object.keys and switch when support is good enough
			for (var key in value)
			{
				if (value.hasOwnProperty(key))
				{
					var result = _Json_runHelp(decoder.b, value[key]);
					if (!elm$core$Result$isOk(result))
					{
						return elm$core$Result$Err(A2(elm$json$Json$Decode$Field, key, result.a));
					}
					keyValuePairs = _List_Cons(_Utils_Tuple2(key, result.a), keyValuePairs);
				}
			}
			return elm$core$Result$Ok(elm$core$List$reverse(keyValuePairs));

		case 9:
			var answer = decoder.f;
			var decoders = decoder.g;
			for (var i = 0; i < decoders.length; i++)
			{
				var result = _Json_runHelp(decoders[i], value);
				if (!elm$core$Result$isOk(result))
				{
					return result;
				}
				answer = answer(result.a);
			}
			return elm$core$Result$Ok(answer);

		case 10:
			var result = _Json_runHelp(decoder.b, value);
			return (!elm$core$Result$isOk(result))
				? result
				: _Json_runHelp(decoder.h(result.a), value);

		case 11:
			var errors = _List_Nil;
			for (var temp = decoder.g; temp.b; temp = temp.b) // WHILE_CONS
			{
				var result = _Json_runHelp(temp.a, value);
				if (elm$core$Result$isOk(result))
				{
					return result;
				}
				errors = _List_Cons(result.a, errors);
			}
			return elm$core$Result$Err(elm$json$Json$Decode$OneOf(elm$core$List$reverse(errors)));

		case 1:
			return elm$core$Result$Err(A2(elm$json$Json$Decode$Failure, decoder.a, _Json_wrap(value)));

		case 0:
			return elm$core$Result$Ok(decoder.a);
	}
}

function _Json_runArrayDecoder(decoder, value, toElmValue)
{
	var len = value.length;
	var array = new Array(len);
	for (var i = 0; i < len; i++)
	{
		var result = _Json_runHelp(decoder, value[i]);
		if (!elm$core$Result$isOk(result))
		{
			return elm$core$Result$Err(A2(elm$json$Json$Decode$Index, i, result.a));
		}
		array[i] = result.a;
	}
	return elm$core$Result$Ok(toElmValue(array));
}

function _Json_isArray(value)
{
	return Array.isArray(value) || value instanceof FileList;
}

function _Json_toElmArray(array)
{
	return A2(elm$core$Array$initialize, array.length, function(i) { return array[i]; });
}

function _Json_expecting(type, value)
{
	return elm$core$Result$Err(A2(elm$json$Json$Decode$Failure, 'Expecting ' + type, _Json_wrap(value)));
}


// EQUALITY

function _Json_equality(x, y)
{
	if (x === y)
	{
		return true;
	}

	if (x.$ !== y.$)
	{
		return false;
	}

	switch (x.$)
	{
		case 0:
		case 1:
			return x.a === y.a;

		case 2:
			return x.b === y.b;

		case 5:
			return x.c === y.c;

		case 3:
		case 4:
		case 8:
			return _Json_equality(x.b, y.b);

		case 6:
			return x.d === y.d && _Json_equality(x.b, y.b);

		case 7:
			return x.e === y.e && _Json_equality(x.b, y.b);

		case 9:
			return x.f === y.f && _Json_listEquality(x.g, y.g);

		case 10:
			return x.h === y.h && _Json_equality(x.b, y.b);

		case 11:
			return _Json_listEquality(x.g, y.g);
	}
}

function _Json_listEquality(aDecoders, bDecoders)
{
	var len = aDecoders.length;
	if (len !== bDecoders.length)
	{
		return false;
	}
	for (var i = 0; i < len; i++)
	{
		if (!_Json_equality(aDecoders[i], bDecoders[i]))
		{
			return false;
		}
	}
	return true;
}


// ENCODE

var _Json_encode = F2(function(indentLevel, value)
{
	return JSON.stringify(_Json_unwrap(value), null, indentLevel) + '';
});

function _Json_wrap(value) { return { $: 0, a: value }; }
function _Json_unwrap(value) { return value.a; }

function _Json_wrap_UNUSED(value) { return value; }
function _Json_unwrap_UNUSED(value) { return value; }

function _Json_emptyArray() { return []; }
function _Json_emptyObject() { return {}; }

var _Json_addField = F3(function(key, value, object)
{
	object[key] = _Json_unwrap(value);
	return object;
});

function _Json_addEntry(func)
{
	return F2(function(entry, array)
	{
		array.push(_Json_unwrap(func(entry)));
		return array;
	});
}

var _Json_encodeNull = _Json_wrap(null);



// TASKS

function _Scheduler_succeed(value)
{
	return {
		$: 0,
		a: value
	};
}

function _Scheduler_fail(error)
{
	return {
		$: 1,
		a: error
	};
}

function _Scheduler_binding(callback)
{
	return {
		$: 2,
		b: callback,
		c: null
	};
}

var _Scheduler_andThen = F2(function(callback, task)
{
	return {
		$: 3,
		b: callback,
		d: task
	};
});

var _Scheduler_onError = F2(function(callback, task)
{
	return {
		$: 4,
		b: callback,
		d: task
	};
});

function _Scheduler_receive(callback)
{
	return {
		$: 5,
		b: callback
	};
}


// PROCESSES

var _Scheduler_guid = 0;

function _Scheduler_rawSpawn(task)
{
	var proc = {
		$: 0,
		e: _Scheduler_guid++,
		f: task,
		g: null,
		h: []
	};

	_Scheduler_enqueue(proc);

	return proc;
}

function _Scheduler_spawn(task)
{
	return _Scheduler_binding(function(callback) {
		callback(_Scheduler_succeed(_Scheduler_rawSpawn(task)));
	});
}

function _Scheduler_rawSend(proc, msg)
{
	proc.h.push(msg);
	_Scheduler_enqueue(proc);
}

var _Scheduler_send = F2(function(proc, msg)
{
	return _Scheduler_binding(function(callback) {
		_Scheduler_rawSend(proc, msg);
		callback(_Scheduler_succeed(_Utils_Tuple0));
	});
});

function _Scheduler_kill(proc)
{
	return _Scheduler_binding(function(callback) {
		var task = proc.f;
		if (task.$ === 2 && task.c)
		{
			task.c();
		}

		proc.f = null;

		callback(_Scheduler_succeed(_Utils_Tuple0));
	});
}


/* STEP PROCESSES

type alias Process =
  { $ : tag
  , id : unique_id
  , root : Task
  , stack : null | { $: SUCCEED | FAIL, a: callback, b: stack }
  , mailbox : [msg]
  }

*/


var _Scheduler_working = false;
var _Scheduler_queue = [];


function _Scheduler_enqueue(proc)
{
	_Scheduler_queue.push(proc);
	if (_Scheduler_working)
	{
		return;
	}
	_Scheduler_working = true;
	while (proc = _Scheduler_queue.shift())
	{
		_Scheduler_step(proc);
	}
	_Scheduler_working = false;
}


function _Scheduler_step(proc)
{
	while (proc.f)
	{
		var rootTag = proc.f.$;
		if (rootTag === 0 || rootTag === 1)
		{
			while (proc.g && proc.g.$ !== rootTag)
			{
				proc.g = proc.g.i;
			}
			if (!proc.g)
			{
				return;
			}
			proc.f = proc.g.b(proc.f.a);
			proc.g = proc.g.i;
		}
		else if (rootTag === 2)
		{
			proc.f.c = proc.f.b(function(newRoot) {
				proc.f = newRoot;
				_Scheduler_enqueue(proc);
			});
			return;
		}
		else if (rootTag === 5)
		{
			if (proc.h.length === 0)
			{
				return;
			}
			proc.f = proc.f.b(proc.h.shift());
		}
		else // if (rootTag === 3 || rootTag === 4)
		{
			proc.g = {
				$: rootTag === 3 ? 0 : 1,
				b: proc.f.b,
				i: proc.g
			};
			proc.f = proc.f.d;
		}
	}
}



function _Process_sleep(time)
{
	return _Scheduler_binding(function(callback) {
		var id = setTimeout(function() {
			callback(_Scheduler_succeed(_Utils_Tuple0));
		}, time);

		return function() { clearTimeout(id); };
	});
}




// PROGRAMS


var _Platform_worker = F4(function(impl, flagDecoder, debugMetadata, args)
{
	return _Platform_initialize(
		flagDecoder,
		args,
		impl.init,
		impl.update,
		impl.subscriptions,
		function() { return function() {} }
	);
});



// INITIALIZE A PROGRAM


function _Platform_initialize(flagDecoder, args, init, update, subscriptions, stepperBuilder)
{
	var result = A2(_Json_run, flagDecoder, _Json_wrap(args ? args['flags'] : undefined));
	elm$core$Result$isOk(result) || _Debug_crash(2 /**/, _Json_errorToString(result.a) /**/);
	var managers = {};
	result = init(result.a);
	var model = result.a;
	var stepper = stepperBuilder(sendToApp, model);
	var ports = _Platform_setupEffects(managers, sendToApp);

	function sendToApp(msg, viewMetadata)
	{
		result = A2(update, msg, model);
		stepper(model = result.a, viewMetadata);
		_Platform_dispatchEffects(managers, result.b, subscriptions(model));
	}

	_Platform_dispatchEffects(managers, result.b, subscriptions(model));

	return ports ? { ports: ports } : {};
}



// TRACK PRELOADS
//
// This is used by code in elm/browser and elm/http
// to register any HTTP requests that are triggered by init.
//


var _Platform_preload;


function _Platform_registerPreload(url)
{
	_Platform_preload.add(url);
}



// EFFECT MANAGERS


var _Platform_effectManagers = {};


function _Platform_setupEffects(managers, sendToApp)
{
	var ports;

	// setup all necessary effect managers
	for (var key in _Platform_effectManagers)
	{
		var manager = _Platform_effectManagers[key];

		if (manager.a)
		{
			ports = ports || {};
			ports[key] = manager.a(key, sendToApp);
		}

		managers[key] = _Platform_instantiateManager(manager, sendToApp);
	}

	return ports;
}


function _Platform_createManager(init, onEffects, onSelfMsg, cmdMap, subMap)
{
	return {
		b: init,
		c: onEffects,
		d: onSelfMsg,
		e: cmdMap,
		f: subMap
	};
}


function _Platform_instantiateManager(info, sendToApp)
{
	var router = {
		g: sendToApp,
		h: undefined
	};

	var onEffects = info.c;
	var onSelfMsg = info.d;
	var cmdMap = info.e;
	var subMap = info.f;

	function loop(state)
	{
		return A2(_Scheduler_andThen, loop, _Scheduler_receive(function(msg)
		{
			var value = msg.a;

			if (msg.$ === 0)
			{
				return A3(onSelfMsg, router, value, state);
			}

			return cmdMap && subMap
				? A4(onEffects, router, value.i, value.j, state)
				: A3(onEffects, router, cmdMap ? value.i : value.j, state);
		}));
	}

	return router.h = _Scheduler_rawSpawn(A2(_Scheduler_andThen, loop, info.b));
}



// ROUTING


var _Platform_sendToApp = F2(function(router, msg)
{
	return _Scheduler_binding(function(callback)
	{
		router.g(msg);
		callback(_Scheduler_succeed(_Utils_Tuple0));
	});
});


var _Platform_sendToSelf = F2(function(router, msg)
{
	return A2(_Scheduler_send, router.h, {
		$: 0,
		a: msg
	});
});



// BAGS


function _Platform_leaf(home)
{
	return function(value)
	{
		return {
			$: 1,
			k: home,
			l: value
		};
	};
}


function _Platform_batch(list)
{
	return {
		$: 2,
		m: list
	};
}


var _Platform_map = F2(function(tagger, bag)
{
	return {
		$: 3,
		n: tagger,
		o: bag
	}
});



// PIPE BAGS INTO EFFECT MANAGERS


function _Platform_dispatchEffects(managers, cmdBag, subBag)
{
	var effectsDict = {};
	_Platform_gatherEffects(true, cmdBag, effectsDict, null);
	_Platform_gatherEffects(false, subBag, effectsDict, null);

	for (var home in managers)
	{
		_Scheduler_rawSend(managers[home], {
			$: 'fx',
			a: effectsDict[home] || { i: _List_Nil, j: _List_Nil }
		});
	}
}


function _Platform_gatherEffects(isCmd, bag, effectsDict, taggers)
{
	switch (bag.$)
	{
		case 1:
			var home = bag.k;
			var effect = _Platform_toEffect(isCmd, home, taggers, bag.l);
			effectsDict[home] = _Platform_insert(isCmd, effect, effectsDict[home]);
			return;

		case 2:
			for (var list = bag.m; list.b; list = list.b) // WHILE_CONS
			{
				_Platform_gatherEffects(isCmd, list.a, effectsDict, taggers);
			}
			return;

		case 3:
			_Platform_gatherEffects(isCmd, bag.o, effectsDict, {
				p: bag.n,
				q: taggers
			});
			return;
	}
}


function _Platform_toEffect(isCmd, home, taggers, value)
{
	function applyTaggers(x)
	{
		for (var temp = taggers; temp; temp = temp.q)
		{
			x = temp.p(x);
		}
		return x;
	}

	var map = isCmd
		? _Platform_effectManagers[home].e
		: _Platform_effectManagers[home].f;

	return A2(map, applyTaggers, value)
}


function _Platform_insert(isCmd, newEffect, effects)
{
	effects = effects || { i: _List_Nil, j: _List_Nil };

	isCmd
		? (effects.i = _List_Cons(newEffect, effects.i))
		: (effects.j = _List_Cons(newEffect, effects.j));

	return effects;
}



// PORTS


function _Platform_checkPortName(name)
{
	if (_Platform_effectManagers[name])
	{
		_Debug_crash(3, name)
	}
}



// OUTGOING PORTS


function _Platform_outgoingPort(name, converter)
{
	_Platform_checkPortName(name);
	_Platform_effectManagers[name] = {
		e: _Platform_outgoingPortMap,
		r: converter,
		a: _Platform_setupOutgoingPort
	};
	return _Platform_leaf(name);
}


var _Platform_outgoingPortMap = F2(function(tagger, value) { return value; });


function _Platform_setupOutgoingPort(name)
{
	var subs = [];
	var converter = _Platform_effectManagers[name].r;

	// CREATE MANAGER

	var init = _Process_sleep(0);

	_Platform_effectManagers[name].b = init;
	_Platform_effectManagers[name].c = F3(function(router, cmdList, state)
	{
		for ( ; cmdList.b; cmdList = cmdList.b) // WHILE_CONS
		{
			// grab a separate reference to subs in case unsubscribe is called
			var currentSubs = subs;
			var value = _Json_unwrap(converter(cmdList.a));
			for (var i = 0; i < currentSubs.length; i++)
			{
				currentSubs[i](value);
			}
		}
		return init;
	});

	// PUBLIC API

	function subscribe(callback)
	{
		subs.push(callback);
	}

	function unsubscribe(callback)
	{
		// copy subs into a new array in case unsubscribe is called within a
		// subscribed callback
		subs = subs.slice();
		var index = subs.indexOf(callback);
		if (index >= 0)
		{
			subs.splice(index, 1);
		}
	}

	return {
		subscribe: subscribe,
		unsubscribe: unsubscribe
	};
}



// INCOMING PORTS


function _Platform_incomingPort(name, converter)
{
	_Platform_checkPortName(name);
	_Platform_effectManagers[name] = {
		f: _Platform_incomingPortMap,
		r: converter,
		a: _Platform_setupIncomingPort
	};
	return _Platform_leaf(name);
}


var _Platform_incomingPortMap = F2(function(tagger, finalTagger)
{
	return function(value)
	{
		return tagger(finalTagger(value));
	};
});


function _Platform_setupIncomingPort(name, sendToApp)
{
	var subs = _List_Nil;
	var converter = _Platform_effectManagers[name].r;

	// CREATE MANAGER

	var init = _Scheduler_succeed(null);

	_Platform_effectManagers[name].b = init;
	_Platform_effectManagers[name].c = F3(function(router, subList, state)
	{
		subs = subList;
		return init;
	});

	// PUBLIC API

	function send(incomingValue)
	{
		var result = A2(_Json_run, converter, _Json_wrap(incomingValue));

		elm$core$Result$isOk(result) || _Debug_crash(4, name, result.a);

		var value = result.a;
		for (var temp = subs; temp.b; temp = temp.b) // WHILE_CONS
		{
			sendToApp(temp.a(value));
		}
	}

	return { send: send };
}



// EXPORT ELM MODULES
//
// Have DEBUG and PROD versions so that we can (1) give nicer errors in
// debug mode and (2) not pay for the bits needed for that in prod mode.
//


function _Platform_export_UNUSED(exports)
{
	scope['Elm']
		? _Platform_mergeExportsProd(scope['Elm'], exports)
		: scope['Elm'] = exports;
}


function _Platform_mergeExportsProd(obj, exports)
{
	for (var name in exports)
	{
		(name in obj)
			? (name == 'init')
				? _Debug_crash(6)
				: _Platform_mergeExportsProd(obj[name], exports[name])
			: (obj[name] = exports[name]);
	}
}


function _Platform_export(exports)
{
	scope['Elm']
		? _Platform_mergeExportsDebug('Elm', scope['Elm'], exports)
		: scope['Elm'] = exports;
}


function _Platform_mergeExportsDebug(moduleName, obj, exports)
{
	for (var name in exports)
	{
		(name in obj)
			? (name == 'init')
				? _Debug_crash(6, moduleName)
				: _Platform_mergeExportsDebug(moduleName + '.' + name, obj[name], exports[name])
			: (obj[name] = exports[name]);
	}
}



// SEND REQUEST

var _Http_toTask = F3(function(router, toTask, request)
{
	return _Scheduler_binding(function(callback)
	{
		function done(response) {
			callback(toTask(request.expect.a(response)));
		}

		var xhr = new XMLHttpRequest();
		xhr.addEventListener('error', function() { done(elm$http$Http$NetworkError_); });
		xhr.addEventListener('timeout', function() { done(elm$http$Http$Timeout_); });
		xhr.addEventListener('load', function() { done(_Http_toResponse(request.expect.b, xhr)); });
		elm$core$Maybe$isJust(request.tracker) && _Http_track(router, xhr, request.tracker.a);

		try {
			xhr.open(request.method, request.url, true);
		} catch (e) {
			return done(elm$http$Http$BadUrl_(request.url));
		}

		_Http_configureRequest(xhr, request);

		request.body.a && xhr.setRequestHeader('Content-Type', request.body.a);
		xhr.send(request.body.b);

		return function() { xhr.c = true; xhr.abort(); };
	});
});


// CONFIGURE

function _Http_configureRequest(xhr, request)
{
	for (var headers = request.headers; headers.b; headers = headers.b) // WHILE_CONS
	{
		xhr.setRequestHeader(headers.a.a, headers.a.b);
	}
	xhr.timeout = request.timeout.a || 0;
	xhr.responseType = request.expect.d;
	xhr.withCredentials = request.allowCookiesFromOtherDomains;
}


// RESPONSES

function _Http_toResponse(toBody, xhr)
{
	return A2(
		200 <= xhr.status && xhr.status < 300 ? elm$http$Http$GoodStatus_ : elm$http$Http$BadStatus_,
		_Http_toMetadata(xhr),
		toBody(xhr.response)
	);
}


// METADATA

function _Http_toMetadata(xhr)
{
	return {
		url: xhr.responseURL,
		statusCode: xhr.status,
		statusText: xhr.statusText,
		headers: _Http_parseHeaders(xhr.getAllResponseHeaders())
	};
}


// HEADERS

function _Http_parseHeaders(rawHeaders)
{
	if (!rawHeaders)
	{
		return elm$core$Dict$empty;
	}

	var headers = elm$core$Dict$empty;
	var headerPairs = rawHeaders.split('\r\n');
	for (var i = headerPairs.length; i--; )
	{
		var headerPair = headerPairs[i];
		var index = headerPair.indexOf(': ');
		if (index > 0)
		{
			var key = headerPair.substring(0, index);
			var value = headerPair.substring(index + 2);

			headers = A3(elm$core$Dict$update, key, function(oldValue) {
				return elm$core$Maybe$Just(elm$core$Maybe$isJust(oldValue)
					? value + ', ' + oldValue.a
					: value
				);
			}, headers);
		}
	}
	return headers;
}


// EXPECT

var _Http_expect = F3(function(type, toBody, toValue)
{
	return {
		$: 0,
		d: type,
		b: toBody,
		a: toValue
	};
});

var _Http_mapExpect = F2(function(func, expect)
{
	return {
		$: 0,
		d: expect.d,
		b: expect.b,
		a: function(x) { return func(expect.a(x)); }
	};
});

function _Http_toDataView(arrayBuffer)
{
	return new DataView(arrayBuffer);
}


// BODY and PARTS

var _Http_emptyBody = { $: 0 };
var _Http_pair = F2(function(a, b) { return { $: 0, a: a, b: b }; });

function _Http_toFormData(parts)
{
	for (var formData = new FormData(); parts.b; parts = parts.b) // WHILE_CONS
	{
		var part = parts.a;
		formData.append(part.a, part.b);
	}
	return formData;
}

var _Http_bytesToBlob = F2(function(mime, bytes)
{
	return new Blob([bytes], { type: mime });
});


// PROGRESS

function _Http_track(router, xhr, tracker)
{
	// TODO check out lengthComputable on loadstart event

	xhr.upload.addEventListener('progress', function(event) {
		if (xhr.c) { return; }
		_Scheduler_rawSpawn(A2(elm$core$Platform$sendToSelf, router, _Utils_Tuple2(tracker, elm$http$Http$Sending({
			sent: event.loaded,
			size: event.total
		}))));
	});
	xhr.addEventListener('progress', function(event) {
		if (xhr.c) { return; }
		_Scheduler_rawSpawn(A2(elm$core$Platform$sendToSelf, router, _Utils_Tuple2(tracker, elm$http$Http$Receiving({
			received: event.loaded,
			size: event.lengthComputable ? elm$core$Maybe$Just(event.total) : elm$core$Maybe$Nothing
		}))));
	});
}


function _Time_now(millisToPosix)
{
	return _Scheduler_binding(function(callback)
	{
		callback(_Scheduler_succeed(millisToPosix(Date.now())));
	});
}

var _Time_setInterval = F2(function(interval, task)
{
	return _Scheduler_binding(function(callback)
	{
		var id = setInterval(function() { _Scheduler_rawSpawn(task); }, interval);
		return function() { clearInterval(id); };
	});
});

function _Time_here()
{
	return _Scheduler_binding(function(callback)
	{
		callback(_Scheduler_succeed(
			A2(elm$time$Time$customZone, -(new Date().getTimezoneOffset()), _List_Nil)
		));
	});
}


function _Time_getZoneName()
{
	return _Scheduler_binding(function(callback)
	{
		try
		{
			var name = elm$time$Time$Name(Intl.DateTimeFormat().resolvedOptions().timeZone);
		}
		catch (e)
		{
			var name = elm$time$Time$Offset(new Date().getTimezoneOffset());
		}
		callback(_Scheduler_succeed(name));
	});
}




// HELPERS


var _VirtualDom_divertHrefToApp;

var _VirtualDom_doc = typeof document !== 'undefined' ? document : {};


function _VirtualDom_appendChild(parent, child)
{
	parent.appendChild(child);
}

var _VirtualDom_init = F4(function(virtualNode, flagDecoder, debugMetadata, args)
{
	// NOTE: this function needs _Platform_export available to work

	/**_UNUSED/
	var node = args['node'];
	//*/
	/**/
	var node = args && args['node'] ? args['node'] : _Debug_crash(0);
	//*/

	node.parentNode.replaceChild(
		_VirtualDom_render(virtualNode, function() {}),
		node
	);

	return {};
});



// TEXT


function _VirtualDom_text(string)
{
	return {
		$: 0,
		a: string
	};
}



// NODE


var _VirtualDom_nodeNS = F2(function(namespace, tag)
{
	return F2(function(factList, kidList)
	{
		for (var kids = [], descendantsCount = 0; kidList.b; kidList = kidList.b) // WHILE_CONS
		{
			var kid = kidList.a;
			descendantsCount += (kid.b || 0);
			kids.push(kid);
		}
		descendantsCount += kids.length;

		return {
			$: 1,
			c: tag,
			d: _VirtualDom_organizeFacts(factList),
			e: kids,
			f: namespace,
			b: descendantsCount
		};
	});
});


var _VirtualDom_node = _VirtualDom_nodeNS(undefined);



// KEYED NODE


var _VirtualDom_keyedNodeNS = F2(function(namespace, tag)
{
	return F2(function(factList, kidList)
	{
		for (var kids = [], descendantsCount = 0; kidList.b; kidList = kidList.b) // WHILE_CONS
		{
			var kid = kidList.a;
			descendantsCount += (kid.b.b || 0);
			kids.push(kid);
		}
		descendantsCount += kids.length;

		return {
			$: 2,
			c: tag,
			d: _VirtualDom_organizeFacts(factList),
			e: kids,
			f: namespace,
			b: descendantsCount
		};
	});
});


var _VirtualDom_keyedNode = _VirtualDom_keyedNodeNS(undefined);



// CUSTOM


function _VirtualDom_custom(factList, model, render, diff)
{
	return {
		$: 3,
		d: _VirtualDom_organizeFacts(factList),
		g: model,
		h: render,
		i: diff
	};
}



// MAP


var _VirtualDom_map = F2(function(tagger, node)
{
	return {
		$: 4,
		j: tagger,
		k: node,
		b: 1 + (node.b || 0)
	};
});



// LAZY


function _VirtualDom_thunk(refs, thunk)
{
	return {
		$: 5,
		l: refs,
		m: thunk,
		k: undefined
	};
}

var _VirtualDom_lazy = F2(function(func, a)
{
	return _VirtualDom_thunk([func, a], function() {
		return func(a);
	});
});

var _VirtualDom_lazy2 = F3(function(func, a, b)
{
	return _VirtualDom_thunk([func, a, b], function() {
		return A2(func, a, b);
	});
});

var _VirtualDom_lazy3 = F4(function(func, a, b, c)
{
	return _VirtualDom_thunk([func, a, b, c], function() {
		return A3(func, a, b, c);
	});
});

var _VirtualDom_lazy4 = F5(function(func, a, b, c, d)
{
	return _VirtualDom_thunk([func, a, b, c, d], function() {
		return A4(func, a, b, c, d);
	});
});

var _VirtualDom_lazy5 = F6(function(func, a, b, c, d, e)
{
	return _VirtualDom_thunk([func, a, b, c, d, e], function() {
		return A5(func, a, b, c, d, e);
	});
});

var _VirtualDom_lazy6 = F7(function(func, a, b, c, d, e, f)
{
	return _VirtualDom_thunk([func, a, b, c, d, e, f], function() {
		return A6(func, a, b, c, d, e, f);
	});
});

var _VirtualDom_lazy7 = F8(function(func, a, b, c, d, e, f, g)
{
	return _VirtualDom_thunk([func, a, b, c, d, e, f, g], function() {
		return A7(func, a, b, c, d, e, f, g);
	});
});

var _VirtualDom_lazy8 = F9(function(func, a, b, c, d, e, f, g, h)
{
	return _VirtualDom_thunk([func, a, b, c, d, e, f, g, h], function() {
		return A8(func, a, b, c, d, e, f, g, h);
	});
});



// FACTS


var _VirtualDom_on = F2(function(key, handler)
{
	return {
		$: 'a0',
		n: key,
		o: handler
	};
});
var _VirtualDom_style = F2(function(key, value)
{
	return {
		$: 'a1',
		n: key,
		o: value
	};
});
var _VirtualDom_property = F2(function(key, value)
{
	return {
		$: 'a2',
		n: key,
		o: value
	};
});
var _VirtualDom_attribute = F2(function(key, value)
{
	return {
		$: 'a3',
		n: key,
		o: value
	};
});
var _VirtualDom_attributeNS = F3(function(namespace, key, value)
{
	return {
		$: 'a4',
		n: key,
		o: { f: namespace, o: value }
	};
});



// XSS ATTACK VECTOR CHECKS


function _VirtualDom_noScript(tag)
{
	return tag == 'script' ? 'p' : tag;
}

function _VirtualDom_noOnOrFormAction(key)
{
	return /^(on|formAction$)/i.test(key) ? 'data-' + key : key;
}

function _VirtualDom_noInnerHtmlOrFormAction(key)
{
	return key == 'innerHTML' || key == 'formAction' ? 'data-' + key : key;
}

function _VirtualDom_noJavaScriptUri_UNUSED(value)
{
	return /^javascript:/i.test(value.replace(/\s/g,'')) ? '' : value;
}

function _VirtualDom_noJavaScriptUri(value)
{
	return /^javascript:/i.test(value.replace(/\s/g,''))
		? 'javascript:alert("This is an XSS vector. Please use ports or web components instead.")'
		: value;
}

function _VirtualDom_noJavaScriptOrHtmlUri_UNUSED(value)
{
	return /^\s*(javascript:|data:text\/html)/i.test(value) ? '' : value;
}

function _VirtualDom_noJavaScriptOrHtmlUri(value)
{
	return /^\s*(javascript:|data:text\/html)/i.test(value)
		? 'javascript:alert("This is an XSS vector. Please use ports or web components instead.")'
		: value;
}



// MAP FACTS


var _VirtualDom_mapAttribute = F2(function(func, attr)
{
	return (attr.$ === 'a0')
		? A2(_VirtualDom_on, attr.n, _VirtualDom_mapHandler(func, attr.o))
		: attr;
});

function _VirtualDom_mapHandler(func, handler)
{
	var tag = elm$virtual_dom$VirtualDom$toHandlerInt(handler);

	// 0 = Normal
	// 1 = MayStopPropagation
	// 2 = MayPreventDefault
	// 3 = Custom

	return {
		$: handler.$,
		a:
			!tag
				? A2(elm$json$Json$Decode$map, func, handler.a)
				:
			A3(elm$json$Json$Decode$map2,
				tag < 3
					? _VirtualDom_mapEventTuple
					: _VirtualDom_mapEventRecord,
				elm$json$Json$Decode$succeed(func),
				handler.a
			)
	};
}

var _VirtualDom_mapEventTuple = F2(function(func, tuple)
{
	return _Utils_Tuple2(func(tuple.a), tuple.b);
});

var _VirtualDom_mapEventRecord = F2(function(func, record)
{
	return {
		message: func(record.message),
		stopPropagation: record.stopPropagation,
		preventDefault: record.preventDefault
	}
});



// ORGANIZE FACTS


function _VirtualDom_organizeFacts(factList)
{
	for (var facts = {}; factList.b; factList = factList.b) // WHILE_CONS
	{
		var entry = factList.a;

		var tag = entry.$;
		var key = entry.n;
		var value = entry.o;

		if (tag === 'a2')
		{
			(key === 'className')
				? _VirtualDom_addClass(facts, key, _Json_unwrap(value))
				: facts[key] = _Json_unwrap(value);

			continue;
		}

		var subFacts = facts[tag] || (facts[tag] = {});
		(tag === 'a3' && key === 'class')
			? _VirtualDom_addClass(subFacts, key, value)
			: subFacts[key] = value;
	}

	return facts;
}

function _VirtualDom_addClass(object, key, newClass)
{
	var classes = object[key];
	object[key] = classes ? classes + ' ' + newClass : newClass;
}



// RENDER


function _VirtualDom_render(vNode, eventNode)
{
	var tag = vNode.$;

	if (tag === 5)
	{
		return _VirtualDom_render(vNode.k || (vNode.k = vNode.m()), eventNode);
	}

	if (tag === 0)
	{
		return _VirtualDom_doc.createTextNode(vNode.a);
	}

	if (tag === 4)
	{
		var subNode = vNode.k;
		var tagger = vNode.j;

		while (subNode.$ === 4)
		{
			typeof tagger !== 'object'
				? tagger = [tagger, subNode.j]
				: tagger.push(subNode.j);

			subNode = subNode.k;
		}

		var subEventRoot = { j: tagger, p: eventNode };
		var domNode = _VirtualDom_render(subNode, subEventRoot);
		domNode.elm_event_node_ref = subEventRoot;
		return domNode;
	}

	if (tag === 3)
	{
		var domNode = vNode.h(vNode.g);
		_VirtualDom_applyFacts(domNode, eventNode, vNode.d);
		return domNode;
	}

	// at this point `tag` must be 1 or 2

	var domNode = vNode.f
		? _VirtualDom_doc.createElementNS(vNode.f, vNode.c)
		: _VirtualDom_doc.createElement(vNode.c);

	if (_VirtualDom_divertHrefToApp && vNode.c == 'a')
	{
		domNode.addEventListener('click', _VirtualDom_divertHrefToApp(domNode));
	}

	_VirtualDom_applyFacts(domNode, eventNode, vNode.d);

	for (var kids = vNode.e, i = 0; i < kids.length; i++)
	{
		_VirtualDom_appendChild(domNode, _VirtualDom_render(tag === 1 ? kids[i] : kids[i].b, eventNode));
	}

	return domNode;
}



// APPLY FACTS


function _VirtualDom_applyFacts(domNode, eventNode, facts)
{
	for (var key in facts)
	{
		var value = facts[key];

		key === 'a1'
			? _VirtualDom_applyStyles(domNode, value)
			:
		key === 'a0'
			? _VirtualDom_applyEvents(domNode, eventNode, value)
			:
		key === 'a3'
			? _VirtualDom_applyAttrs(domNode, value)
			:
		key === 'a4'
			? _VirtualDom_applyAttrsNS(domNode, value)
			:
		(key !== 'value' || key !== 'checked' || domNode[key] !== value) && (domNode[key] = value);
	}
}



// APPLY STYLES


function _VirtualDom_applyStyles(domNode, styles)
{
	var domNodeStyle = domNode.style;

	for (var key in styles)
	{
		domNodeStyle[key] = styles[key];
	}
}



// APPLY ATTRS


function _VirtualDom_applyAttrs(domNode, attrs)
{
	for (var key in attrs)
	{
		var value = attrs[key];
		value
			? domNode.setAttribute(key, value)
			: domNode.removeAttribute(key);
	}
}



// APPLY NAMESPACED ATTRS


function _VirtualDom_applyAttrsNS(domNode, nsAttrs)
{
	for (var key in nsAttrs)
	{
		var pair = nsAttrs[key];
		var namespace = pair.f;
		var value = pair.o;

		value
			? domNode.setAttributeNS(namespace, key, value)
			: domNode.removeAttributeNS(namespace, key);
	}
}



// APPLY EVENTS


function _VirtualDom_applyEvents(domNode, eventNode, events)
{
	var allCallbacks = domNode.elmFs || (domNode.elmFs = {});

	for (var key in events)
	{
		var newHandler = events[key];
		var oldCallback = allCallbacks[key];

		if (!newHandler)
		{
			domNode.removeEventListener(key, oldCallback);
			allCallbacks[key] = undefined;
			continue;
		}

		if (oldCallback)
		{
			var oldHandler = oldCallback.q;
			if (oldHandler.$ === newHandler.$)
			{
				oldCallback.q = newHandler;
				continue;
			}
			domNode.removeEventListener(key, oldCallback);
		}

		oldCallback = _VirtualDom_makeCallback(eventNode, newHandler);
		domNode.addEventListener(key, oldCallback,
			_VirtualDom_passiveSupported
			&& { passive: elm$virtual_dom$VirtualDom$toHandlerInt(newHandler) < 2 }
		);
		allCallbacks[key] = oldCallback;
	}
}



// PASSIVE EVENTS


var _VirtualDom_passiveSupported;

try
{
	window.addEventListener('t', null, Object.defineProperty({}, 'passive', {
		get: function() { _VirtualDom_passiveSupported = true; }
	}));
}
catch(e) {}



// EVENT HANDLERS


function _VirtualDom_makeCallback(eventNode, initialHandler)
{
	function callback(event)
	{
		var handler = callback.q;
		var result = _Json_runHelp(handler.a, event);

		if (!elm$core$Result$isOk(result))
		{
			return;
		}

		var tag = elm$virtual_dom$VirtualDom$toHandlerInt(handler);

		// 0 = Normal
		// 1 = MayStopPropagation
		// 2 = MayPreventDefault
		// 3 = Custom

		var value = result.a;
		var message = !tag ? value : tag < 3 ? value.a : value.message;
		var stopPropagation = tag == 1 ? value.b : tag == 3 && value.stopPropagation;
		var currentEventNode = (
			stopPropagation && event.stopPropagation(),
			(tag == 2 ? value.b : tag == 3 && value.preventDefault) && event.preventDefault(),
			eventNode
		);
		var tagger;
		var i;
		while (tagger = currentEventNode.j)
		{
			if (typeof tagger == 'function')
			{
				message = tagger(message);
			}
			else
			{
				for (var i = tagger.length; i--; )
				{
					message = tagger[i](message);
				}
			}
			currentEventNode = currentEventNode.p;
		}
		currentEventNode(message, stopPropagation); // stopPropagation implies isSync
	}

	callback.q = initialHandler;

	return callback;
}

function _VirtualDom_equalEvents(x, y)
{
	return x.$ == y.$ && _Json_equality(x.a, y.a);
}



// DIFF


// TODO: Should we do patches like in iOS?
//
// type Patch
//   = At Int Patch
//   | Batch (List Patch)
//   | Change ...
//
// How could it not be better?
//
function _VirtualDom_diff(x, y)
{
	var patches = [];
	_VirtualDom_diffHelp(x, y, patches, 0);
	return patches;
}


function _VirtualDom_pushPatch(patches, type, index, data)
{
	var patch = {
		$: type,
		r: index,
		s: data,
		t: undefined,
		u: undefined
	};
	patches.push(patch);
	return patch;
}


function _VirtualDom_diffHelp(x, y, patches, index)
{
	if (x === y)
	{
		return;
	}

	var xType = x.$;
	var yType = y.$;

	// Bail if you run into different types of nodes. Implies that the
	// structure has changed significantly and it's not worth a diff.
	if (xType !== yType)
	{
		if (xType === 1 && yType === 2)
		{
			y = _VirtualDom_dekey(y);
			yType = 1;
		}
		else
		{
			_VirtualDom_pushPatch(patches, 0, index, y);
			return;
		}
	}

	// Now we know that both nodes are the same $.
	switch (yType)
	{
		case 5:
			var xRefs = x.l;
			var yRefs = y.l;
			var i = xRefs.length;
			var same = i === yRefs.length;
			while (same && i--)
			{
				same = xRefs[i] === yRefs[i];
			}
			if (same)
			{
				y.k = x.k;
				return;
			}
			y.k = y.m();
			var subPatches = [];
			_VirtualDom_diffHelp(x.k, y.k, subPatches, 0);
			subPatches.length > 0 && _VirtualDom_pushPatch(patches, 1, index, subPatches);
			return;

		case 4:
			// gather nested taggers
			var xTaggers = x.j;
			var yTaggers = y.j;
			var nesting = false;

			var xSubNode = x.k;
			while (xSubNode.$ === 4)
			{
				nesting = true;

				typeof xTaggers !== 'object'
					? xTaggers = [xTaggers, xSubNode.j]
					: xTaggers.push(xSubNode.j);

				xSubNode = xSubNode.k;
			}

			var ySubNode = y.k;
			while (ySubNode.$ === 4)
			{
				nesting = true;

				typeof yTaggers !== 'object'
					? yTaggers = [yTaggers, ySubNode.j]
					: yTaggers.push(ySubNode.j);

				ySubNode = ySubNode.k;
			}

			// Just bail if different numbers of taggers. This implies the
			// structure of the virtual DOM has changed.
			if (nesting && xTaggers.length !== yTaggers.length)
			{
				_VirtualDom_pushPatch(patches, 0, index, y);
				return;
			}

			// check if taggers are "the same"
			if (nesting ? !_VirtualDom_pairwiseRefEqual(xTaggers, yTaggers) : xTaggers !== yTaggers)
			{
				_VirtualDom_pushPatch(patches, 2, index, yTaggers);
			}

			// diff everything below the taggers
			_VirtualDom_diffHelp(xSubNode, ySubNode, patches, index + 1);
			return;

		case 0:
			if (x.a !== y.a)
			{
				_VirtualDom_pushPatch(patches, 3, index, y.a);
			}
			return;

		case 1:
			_VirtualDom_diffNodes(x, y, patches, index, _VirtualDom_diffKids);
			return;

		case 2:
			_VirtualDom_diffNodes(x, y, patches, index, _VirtualDom_diffKeyedKids);
			return;

		case 3:
			if (x.h !== y.h)
			{
				_VirtualDom_pushPatch(patches, 0, index, y);
				return;
			}

			var factsDiff = _VirtualDom_diffFacts(x.d, y.d);
			factsDiff && _VirtualDom_pushPatch(patches, 4, index, factsDiff);

			var patch = y.i(x.g, y.g);
			patch && _VirtualDom_pushPatch(patches, 5, index, patch);

			return;
	}
}

// assumes the incoming arrays are the same length
function _VirtualDom_pairwiseRefEqual(as, bs)
{
	for (var i = 0; i < as.length; i++)
	{
		if (as[i] !== bs[i])
		{
			return false;
		}
	}

	return true;
}

function _VirtualDom_diffNodes(x, y, patches, index, diffKids)
{
	// Bail if obvious indicators have changed. Implies more serious
	// structural changes such that it's not worth it to diff.
	if (x.c !== y.c || x.f !== y.f)
	{
		_VirtualDom_pushPatch(patches, 0, index, y);
		return;
	}

	var factsDiff = _VirtualDom_diffFacts(x.d, y.d);
	factsDiff && _VirtualDom_pushPatch(patches, 4, index, factsDiff);

	diffKids(x, y, patches, index);
}



// DIFF FACTS


// TODO Instead of creating a new diff object, it's possible to just test if
// there *is* a diff. During the actual patch, do the diff again and make the
// modifications directly. This way, there's no new allocations. Worth it?
function _VirtualDom_diffFacts(x, y, category)
{
	var diff;

	// look for changes and removals
	for (var xKey in x)
	{
		if (xKey === 'a1' || xKey === 'a0' || xKey === 'a3' || xKey === 'a4')
		{
			var subDiff = _VirtualDom_diffFacts(x[xKey], y[xKey] || {}, xKey);
			if (subDiff)
			{
				diff = diff || {};
				diff[xKey] = subDiff;
			}
			continue;
		}

		// remove if not in the new facts
		if (!(xKey in y))
		{
			diff = diff || {};
			diff[xKey] =
				!category
					? (typeof x[xKey] === 'string' ? '' : null)
					:
				(category === 'a1')
					? ''
					:
				(category === 'a0' || category === 'a3')
					? undefined
					:
				{ f: x[xKey].f, o: undefined };

			continue;
		}

		var xValue = x[xKey];
		var yValue = y[xKey];

		// reference equal, so don't worry about it
		if (xValue === yValue && xKey !== 'value' && xKey !== 'checked'
			|| category === 'a0' && _VirtualDom_equalEvents(xValue, yValue))
		{
			continue;
		}

		diff = diff || {};
		diff[xKey] = yValue;
	}

	// add new stuff
	for (var yKey in y)
	{
		if (!(yKey in x))
		{
			diff = diff || {};
			diff[yKey] = y[yKey];
		}
	}

	return diff;
}



// DIFF KIDS


function _VirtualDom_diffKids(xParent, yParent, patches, index)
{
	var xKids = xParent.e;
	var yKids = yParent.e;

	var xLen = xKids.length;
	var yLen = yKids.length;

	// FIGURE OUT IF THERE ARE INSERTS OR REMOVALS

	if (xLen > yLen)
	{
		_VirtualDom_pushPatch(patches, 6, index, {
			v: yLen,
			i: xLen - yLen
		});
	}
	else if (xLen < yLen)
	{
		_VirtualDom_pushPatch(patches, 7, index, {
			v: xLen,
			e: yKids
		});
	}

	// PAIRWISE DIFF EVERYTHING ELSE

	for (var minLen = xLen < yLen ? xLen : yLen, i = 0; i < minLen; i++)
	{
		var xKid = xKids[i];
		_VirtualDom_diffHelp(xKid, yKids[i], patches, ++index);
		index += xKid.b || 0;
	}
}



// KEYED DIFF


function _VirtualDom_diffKeyedKids(xParent, yParent, patches, rootIndex)
{
	var localPatches = [];

	var changes = {}; // Dict String Entry
	var inserts = []; // Array { index : Int, entry : Entry }
	// type Entry = { tag : String, vnode : VNode, index : Int, data : _ }

	var xKids = xParent.e;
	var yKids = yParent.e;
	var xLen = xKids.length;
	var yLen = yKids.length;
	var xIndex = 0;
	var yIndex = 0;

	var index = rootIndex;

	while (xIndex < xLen && yIndex < yLen)
	{
		var x = xKids[xIndex];
		var y = yKids[yIndex];

		var xKey = x.a;
		var yKey = y.a;
		var xNode = x.b;
		var yNode = y.b;

		// check if keys match

		if (xKey === yKey)
		{
			index++;
			_VirtualDom_diffHelp(xNode, yNode, localPatches, index);
			index += xNode.b || 0;

			xIndex++;
			yIndex++;
			continue;
		}

		// look ahead 1 to detect insertions and removals.

		var xNext = xKids[xIndex + 1];
		var yNext = yKids[yIndex + 1];

		if (xNext)
		{
			var xNextKey = xNext.a;
			var xNextNode = xNext.b;
			var oldMatch = yKey === xNextKey;
		}

		if (yNext)
		{
			var yNextKey = yNext.a;
			var yNextNode = yNext.b;
			var newMatch = xKey === yNextKey;
		}


		// swap x and y
		if (newMatch && oldMatch)
		{
			index++;
			_VirtualDom_diffHelp(xNode, yNextNode, localPatches, index);
			_VirtualDom_insertNode(changes, localPatches, xKey, yNode, yIndex, inserts);
			index += xNode.b || 0;

			index++;
			_VirtualDom_removeNode(changes, localPatches, xKey, xNextNode, index);
			index += xNextNode.b || 0;

			xIndex += 2;
			yIndex += 2;
			continue;
		}

		// insert y
		if (newMatch)
		{
			index++;
			_VirtualDom_insertNode(changes, localPatches, yKey, yNode, yIndex, inserts);
			_VirtualDom_diffHelp(xNode, yNextNode, localPatches, index);
			index += xNode.b || 0;

			xIndex += 1;
			yIndex += 2;
			continue;
		}

		// remove x
		if (oldMatch)
		{
			index++;
			_VirtualDom_removeNode(changes, localPatches, xKey, xNode, index);
			index += xNode.b || 0;

			index++;
			_VirtualDom_diffHelp(xNextNode, yNode, localPatches, index);
			index += xNextNode.b || 0;

			xIndex += 2;
			yIndex += 1;
			continue;
		}

		// remove x, insert y
		if (xNext && xNextKey === yNextKey)
		{
			index++;
			_VirtualDom_removeNode(changes, localPatches, xKey, xNode, index);
			_VirtualDom_insertNode(changes, localPatches, yKey, yNode, yIndex, inserts);
			index += xNode.b || 0;

			index++;
			_VirtualDom_diffHelp(xNextNode, yNextNode, localPatches, index);
			index += xNextNode.b || 0;

			xIndex += 2;
			yIndex += 2;
			continue;
		}

		break;
	}

	// eat up any remaining nodes with removeNode and insertNode

	while (xIndex < xLen)
	{
		index++;
		var x = xKids[xIndex];
		var xNode = x.b;
		_VirtualDom_removeNode(changes, localPatches, x.a, xNode, index);
		index += xNode.b || 0;
		xIndex++;
	}

	while (yIndex < yLen)
	{
		var endInserts = endInserts || [];
		var y = yKids[yIndex];
		_VirtualDom_insertNode(changes, localPatches, y.a, y.b, undefined, endInserts);
		yIndex++;
	}

	if (localPatches.length > 0 || inserts.length > 0 || endInserts)
	{
		_VirtualDom_pushPatch(patches, 8, rootIndex, {
			w: localPatches,
			x: inserts,
			y: endInserts
		});
	}
}



// CHANGES FROM KEYED DIFF


var _VirtualDom_POSTFIX = '_elmW6BL';


function _VirtualDom_insertNode(changes, localPatches, key, vnode, yIndex, inserts)
{
	var entry = changes[key];

	// never seen this key before
	if (!entry)
	{
		entry = {
			c: 0,
			z: vnode,
			r: yIndex,
			s: undefined
		};

		inserts.push({ r: yIndex, A: entry });
		changes[key] = entry;

		return;
	}

	// this key was removed earlier, a match!
	if (entry.c === 1)
	{
		inserts.push({ r: yIndex, A: entry });

		entry.c = 2;
		var subPatches = [];
		_VirtualDom_diffHelp(entry.z, vnode, subPatches, entry.r);
		entry.r = yIndex;
		entry.s.s = {
			w: subPatches,
			A: entry
		};

		return;
	}

	// this key has already been inserted or moved, a duplicate!
	_VirtualDom_insertNode(changes, localPatches, key + _VirtualDom_POSTFIX, vnode, yIndex, inserts);
}


function _VirtualDom_removeNode(changes, localPatches, key, vnode, index)
{
	var entry = changes[key];

	// never seen this key before
	if (!entry)
	{
		var patch = _VirtualDom_pushPatch(localPatches, 9, index, undefined);

		changes[key] = {
			c: 1,
			z: vnode,
			r: index,
			s: patch
		};

		return;
	}

	// this key was inserted earlier, a match!
	if (entry.c === 0)
	{
		entry.c = 2;
		var subPatches = [];
		_VirtualDom_diffHelp(vnode, entry.z, subPatches, index);

		_VirtualDom_pushPatch(localPatches, 9, index, {
			w: subPatches,
			A: entry
		});

		return;
	}

	// this key has already been removed or moved, a duplicate!
	_VirtualDom_removeNode(changes, localPatches, key + _VirtualDom_POSTFIX, vnode, index);
}



// ADD DOM NODES
//
// Each DOM node has an "index" assigned in order of traversal. It is important
// to minimize our crawl over the actual DOM, so these indexes (along with the
// descendantsCount of virtual nodes) let us skip touching entire subtrees of
// the DOM if we know there are no patches there.


function _VirtualDom_addDomNodes(domNode, vNode, patches, eventNode)
{
	_VirtualDom_addDomNodesHelp(domNode, vNode, patches, 0, 0, vNode.b, eventNode);
}


// assumes `patches` is non-empty and indexes increase monotonically.
function _VirtualDom_addDomNodesHelp(domNode, vNode, patches, i, low, high, eventNode)
{
	var patch = patches[i];
	var index = patch.r;

	while (index === low)
	{
		var patchType = patch.$;

		if (patchType === 1)
		{
			_VirtualDom_addDomNodes(domNode, vNode.k, patch.s, eventNode);
		}
		else if (patchType === 8)
		{
			patch.t = domNode;
			patch.u = eventNode;

			var subPatches = patch.s.w;
			if (subPatches.length > 0)
			{
				_VirtualDom_addDomNodesHelp(domNode, vNode, subPatches, 0, low, high, eventNode);
			}
		}
		else if (patchType === 9)
		{
			patch.t = domNode;
			patch.u = eventNode;

			var data = patch.s;
			if (data)
			{
				data.A.s = domNode;
				var subPatches = data.w;
				if (subPatches.length > 0)
				{
					_VirtualDom_addDomNodesHelp(domNode, vNode, subPatches, 0, low, high, eventNode);
				}
			}
		}
		else
		{
			patch.t = domNode;
			patch.u = eventNode;
		}

		i++;

		if (!(patch = patches[i]) || (index = patch.r) > high)
		{
			return i;
		}
	}

	var tag = vNode.$;

	if (tag === 4)
	{
		var subNode = vNode.k;

		while (subNode.$ === 4)
		{
			subNode = subNode.k;
		}

		return _VirtualDom_addDomNodesHelp(domNode, subNode, patches, i, low + 1, high, domNode.elm_event_node_ref);
	}

	// tag must be 1 or 2 at this point

	var vKids = vNode.e;
	var childNodes = domNode.childNodes;
	for (var j = 0; j < vKids.length; j++)
	{
		low++;
		var vKid = tag === 1 ? vKids[j] : vKids[j].b;
		var nextLow = low + (vKid.b || 0);
		if (low <= index && index <= nextLow)
		{
			i = _VirtualDom_addDomNodesHelp(childNodes[j], vKid, patches, i, low, nextLow, eventNode);
			if (!(patch = patches[i]) || (index = patch.r) > high)
			{
				return i;
			}
		}
		low = nextLow;
	}
	return i;
}



// APPLY PATCHES


function _VirtualDom_applyPatches(rootDomNode, oldVirtualNode, patches, eventNode)
{
	if (patches.length === 0)
	{
		return rootDomNode;
	}

	_VirtualDom_addDomNodes(rootDomNode, oldVirtualNode, patches, eventNode);
	return _VirtualDom_applyPatchesHelp(rootDomNode, patches);
}

function _VirtualDom_applyPatchesHelp(rootDomNode, patches)
{
	for (var i = 0; i < patches.length; i++)
	{
		var patch = patches[i];
		var localDomNode = patch.t
		var newNode = _VirtualDom_applyPatch(localDomNode, patch);
		if (localDomNode === rootDomNode)
		{
			rootDomNode = newNode;
		}
	}
	return rootDomNode;
}

function _VirtualDom_applyPatch(domNode, patch)
{
	switch (patch.$)
	{
		case 0:
			return _VirtualDom_applyPatchRedraw(domNode, patch.s, patch.u);

		case 4:
			_VirtualDom_applyFacts(domNode, patch.u, patch.s);
			return domNode;

		case 3:
			domNode.replaceData(0, domNode.length, patch.s);
			return domNode;

		case 1:
			return _VirtualDom_applyPatchesHelp(domNode, patch.s);

		case 2:
			if (domNode.elm_event_node_ref)
			{
				domNode.elm_event_node_ref.j = patch.s;
			}
			else
			{
				domNode.elm_event_node_ref = { j: patch.s, p: patch.u };
			}
			return domNode;

		case 6:
			var data = patch.s;
			for (var i = 0; i < data.i; i++)
			{
				domNode.removeChild(domNode.childNodes[data.v]);
			}
			return domNode;

		case 7:
			var data = patch.s;
			var kids = data.e;
			var i = data.v;
			var theEnd = domNode.childNodes[i];
			for (; i < kids.length; i++)
			{
				domNode.insertBefore(_VirtualDom_render(kids[i], patch.u), theEnd);
			}
			return domNode;

		case 9:
			var data = patch.s;
			if (!data)
			{
				domNode.parentNode.removeChild(domNode);
				return domNode;
			}
			var entry = data.A;
			if (typeof entry.r !== 'undefined')
			{
				domNode.parentNode.removeChild(domNode);
			}
			entry.s = _VirtualDom_applyPatchesHelp(domNode, data.w);
			return domNode;

		case 8:
			return _VirtualDom_applyPatchReorder(domNode, patch);

		case 5:
			return patch.s(domNode);

		default:
			_Debug_crash(10); // 'Ran into an unknown patch!'
	}
}


function _VirtualDom_applyPatchRedraw(domNode, vNode, eventNode)
{
	var parentNode = domNode.parentNode;
	var newNode = _VirtualDom_render(vNode, eventNode);

	if (!newNode.elm_event_node_ref)
	{
		newNode.elm_event_node_ref = domNode.elm_event_node_ref;
	}

	if (parentNode && newNode !== domNode)
	{
		parentNode.replaceChild(newNode, domNode);
	}
	return newNode;
}


function _VirtualDom_applyPatchReorder(domNode, patch)
{
	var data = patch.s;

	// remove end inserts
	var frag = _VirtualDom_applyPatchReorderEndInsertsHelp(data.y, patch);

	// removals
	domNode = _VirtualDom_applyPatchesHelp(domNode, data.w);

	// inserts
	var inserts = data.x;
	for (var i = 0; i < inserts.length; i++)
	{
		var insert = inserts[i];
		var entry = insert.A;
		var node = entry.c === 2
			? entry.s
			: _VirtualDom_render(entry.z, patch.u);
		domNode.insertBefore(node, domNode.childNodes[insert.r]);
	}

	// add end inserts
	if (frag)
	{
		_VirtualDom_appendChild(domNode, frag);
	}

	return domNode;
}


function _VirtualDom_applyPatchReorderEndInsertsHelp(endInserts, patch)
{
	if (!endInserts)
	{
		return;
	}

	var frag = _VirtualDom_doc.createDocumentFragment();
	for (var i = 0; i < endInserts.length; i++)
	{
		var insert = endInserts[i];
		var entry = insert.A;
		_VirtualDom_appendChild(frag, entry.c === 2
			? entry.s
			: _VirtualDom_render(entry.z, patch.u)
		);
	}
	return frag;
}


function _VirtualDom_virtualize(node)
{
	// TEXT NODES

	if (node.nodeType === 3)
	{
		return _VirtualDom_text(node.textContent);
	}


	// WEIRD NODES

	if (node.nodeType !== 1)
	{
		return _VirtualDom_text('');
	}


	// ELEMENT NODES

	var attrList = _List_Nil;
	var attrs = node.attributes;
	for (var i = attrs.length; i--; )
	{
		var attr = attrs[i];
		var name = attr.name;
		var value = attr.value;
		attrList = _List_Cons( A2(_VirtualDom_attribute, name, value), attrList );
	}

	var tag = node.tagName.toLowerCase();
	var kidList = _List_Nil;
	var kids = node.childNodes;

	for (var i = kids.length; i--; )
	{
		kidList = _List_Cons(_VirtualDom_virtualize(kids[i]), kidList);
	}
	return A3(_VirtualDom_node, tag, attrList, kidList);
}

function _VirtualDom_dekey(keyedNode)
{
	var keyedKids = keyedNode.e;
	var len = keyedKids.length;
	var kids = new Array(len);
	for (var i = 0; i < len; i++)
	{
		kids[i] = keyedKids[i].b;
	}

	return {
		$: 1,
		c: keyedNode.c,
		d: keyedNode.d,
		e: kids,
		f: keyedNode.f,
		b: keyedNode.b
	};
}



// ELEMENT


var _Debugger_element;

var _Browser_element = _Debugger_element || F4(function(impl, flagDecoder, debugMetadata, args)
{
	return _Platform_initialize(
		flagDecoder,
		args,
		impl.init,
		impl.update,
		impl.subscriptions,
		function(sendToApp, initialModel) {
			var view = impl.view;
			/**_UNUSED/
			var domNode = args['node'];
			//*/
			/**/
			var domNode = args && args['node'] ? args['node'] : _Debug_crash(0);
			//*/
			var currNode = _VirtualDom_virtualize(domNode);

			return _Browser_makeAnimator(initialModel, function(model)
			{
				var nextNode = view(model);
				var patches = _VirtualDom_diff(currNode, nextNode);
				domNode = _VirtualDom_applyPatches(domNode, currNode, patches, sendToApp);
				currNode = nextNode;
			});
		}
	);
});



// DOCUMENT


var _Debugger_document;

var _Browser_document = _Debugger_document || F4(function(impl, flagDecoder, debugMetadata, args)
{
	return _Platform_initialize(
		flagDecoder,
		args,
		impl.init,
		impl.update,
		impl.subscriptions,
		function(sendToApp, initialModel) {
			var divertHrefToApp = impl.setup && impl.setup(sendToApp)
			var view = impl.view;
			var title = _VirtualDom_doc.title;
			var bodyNode = _VirtualDom_doc.body;
			var currNode = _VirtualDom_virtualize(bodyNode);
			return _Browser_makeAnimator(initialModel, function(model)
			{
				_VirtualDom_divertHrefToApp = divertHrefToApp;
				var doc = view(model);
				var nextNode = _VirtualDom_node('body')(_List_Nil)(doc.body);
				var patches = _VirtualDom_diff(currNode, nextNode);
				bodyNode = _VirtualDom_applyPatches(bodyNode, currNode, patches, sendToApp);
				currNode = nextNode;
				_VirtualDom_divertHrefToApp = 0;
				(title !== doc.title) && (_VirtualDom_doc.title = title = doc.title);
			});
		}
	);
});



// ANIMATION


var _Browser_requestAnimationFrame =
	typeof requestAnimationFrame !== 'undefined'
		? requestAnimationFrame
		: function(callback) { setTimeout(callback, 1000 / 60); };


function _Browser_makeAnimator(model, draw)
{
	draw(model);

	var state = 0;

	function updateIfNeeded()
	{
		state = state === 1
			? 0
			: ( _Browser_requestAnimationFrame(updateIfNeeded), draw(model), 1 );
	}

	return function(nextModel, isSync)
	{
		model = nextModel;

		isSync
			? ( draw(model),
				state === 2 && (state = 1)
				)
			: ( state === 0 && _Browser_requestAnimationFrame(updateIfNeeded),
				state = 2
				);
	};
}



// APPLICATION


function _Browser_application(impl)
{
	var onUrlChange = impl.onUrlChange;
	var onUrlRequest = impl.onUrlRequest;
	var key = function() { key.a(onUrlChange(_Browser_getUrl())); };

	return _Browser_document({
		setup: function(sendToApp)
		{
			key.a = sendToApp;
			_Browser_window.addEventListener('popstate', key);
			_Browser_window.navigator.userAgent.indexOf('Trident') < 0 || _Browser_window.addEventListener('hashchange', key);

			return F2(function(domNode, event)
			{
				if (!event.ctrlKey && !event.metaKey && !event.shiftKey && event.button < 1 && !domNode.target && !domNode.download)
				{
					event.preventDefault();
					var href = domNode.href;
					var curr = _Browser_getUrl();
					var next = elm$url$Url$fromString(href).a;
					sendToApp(onUrlRequest(
						(next
							&& curr.protocol === next.protocol
							&& curr.host === next.host
							&& curr.port_.a === next.port_.a
						)
							? elm$browser$Browser$Internal(next)
							: elm$browser$Browser$External(href)
					));
				}
			});
		},
		init: function(flags)
		{
			return A3(impl.init, flags, _Browser_getUrl(), key);
		},
		view: impl.view,
		update: impl.update,
		subscriptions: impl.subscriptions
	});
}

function _Browser_getUrl()
{
	return elm$url$Url$fromString(_VirtualDom_doc.location.href).a || _Debug_crash(1);
}

var _Browser_go = F2(function(key, n)
{
	return A2(elm$core$Task$perform, elm$core$Basics$never, _Scheduler_binding(function() {
		n && history.go(n);
		key();
	}));
});

var _Browser_pushUrl = F2(function(key, url)
{
	return A2(elm$core$Task$perform, elm$core$Basics$never, _Scheduler_binding(function() {
		history.pushState({}, '', url);
		key();
	}));
});

var _Browser_replaceUrl = F2(function(key, url)
{
	return A2(elm$core$Task$perform, elm$core$Basics$never, _Scheduler_binding(function() {
		history.replaceState({}, '', url);
		key();
	}));
});



// GLOBAL EVENTS


var _Browser_fakeNode = { addEventListener: function() {}, removeEventListener: function() {} };
var _Browser_doc = typeof document !== 'undefined' ? document : _Browser_fakeNode;
var _Browser_window = typeof window !== 'undefined' ? window : _Browser_fakeNode;

var _Browser_on = F3(function(node, eventName, sendToSelf)
{
	return _Scheduler_spawn(_Scheduler_binding(function(callback)
	{
		function handler(event)	{ _Scheduler_rawSpawn(sendToSelf(event)); }
		node.addEventListener(eventName, handler, _VirtualDom_passiveSupported && { passive: true });
		return function() { node.removeEventListener(eventName, handler); };
	}));
});

var _Browser_decodeEvent = F2(function(decoder, event)
{
	var result = _Json_runHelp(decoder, event);
	return elm$core$Result$isOk(result) ? elm$core$Maybe$Just(result.a) : elm$core$Maybe$Nothing;
});



// PAGE VISIBILITY


function _Browser_visibilityInfo()
{
	return (typeof _VirtualDom_doc.hidden !== 'undefined')
		? { hidden: 'hidden', change: 'visibilitychange' }
		:
	(typeof _VirtualDom_doc.mozHidden !== 'undefined')
		? { hidden: 'mozHidden', change: 'mozvisibilitychange' }
		:
	(typeof _VirtualDom_doc.msHidden !== 'undefined')
		? { hidden: 'msHidden', change: 'msvisibilitychange' }
		:
	(typeof _VirtualDom_doc.webkitHidden !== 'undefined')
		? { hidden: 'webkitHidden', change: 'webkitvisibilitychange' }
		: { hidden: 'hidden', change: 'visibilitychange' };
}



// ANIMATION FRAMES


function _Browser_rAF()
{
	return _Scheduler_binding(function(callback)
	{
		var id = requestAnimationFrame(function() {
			callback(_Scheduler_succeed(Date.now()));
		});

		return function() {
			cancelAnimationFrame(id);
		};
	});
}


function _Browser_now()
{
	return _Scheduler_binding(function(callback)
	{
		callback(_Scheduler_succeed(Date.now()));
	});
}



// DOM STUFF


function _Browser_withNode(id, doStuff)
{
	return _Scheduler_binding(function(callback)
	{
		_Browser_requestAnimationFrame(function() {
			var node = document.getElementById(id);
			callback(node
				? _Scheduler_succeed(doStuff(node))
				: _Scheduler_fail(elm$browser$Browser$Dom$NotFound(id))
			);
		});
	});
}


function _Browser_withWindow(doStuff)
{
	return _Scheduler_binding(function(callback)
	{
		_Browser_requestAnimationFrame(function() {
			callback(_Scheduler_succeed(doStuff()));
		});
	});
}


// FOCUS and BLUR


var _Browser_call = F2(function(functionName, id)
{
	return _Browser_withNode(id, function(node) {
		node[functionName]();
		return _Utils_Tuple0;
	});
});



// WINDOW VIEWPORT


function _Browser_getViewport()
{
	return {
		scene: _Browser_getScene(),
		viewport: {
			x: _Browser_window.pageXOffset,
			y: _Browser_window.pageYOffset,
			width: _Browser_doc.documentElement.clientWidth,
			height: _Browser_doc.documentElement.clientHeight
		}
	};
}

function _Browser_getScene()
{
	var body = _Browser_doc.body;
	var elem = _Browser_doc.documentElement;
	return {
		width: Math.max(body.scrollWidth, body.offsetWidth, elem.scrollWidth, elem.offsetWidth, elem.clientWidth),
		height: Math.max(body.scrollHeight, body.offsetHeight, elem.scrollHeight, elem.offsetHeight, elem.clientHeight)
	};
}

var _Browser_setViewport = F2(function(x, y)
{
	return _Browser_withWindow(function()
	{
		_Browser_window.scroll(x, y);
		return _Utils_Tuple0;
	});
});



// ELEMENT VIEWPORT


function _Browser_getViewportOf(id)
{
	return _Browser_withNode(id, function(node)
	{
		return {
			scene: {
				width: node.scrollWidth,
				height: node.scrollHeight
			},
			viewport: {
				x: node.scrollLeft,
				y: node.scrollTop,
				width: node.clientWidth,
				height: node.clientHeight
			}
		};
	});
}


var _Browser_setViewportOf = F3(function(id, x, y)
{
	return _Browser_withNode(id, function(node)
	{
		node.scrollLeft = x;
		node.scrollTop = y;
		return _Utils_Tuple0;
	});
});



// ELEMENT


function _Browser_getElement(id)
{
	return _Browser_withNode(id, function(node)
	{
		var rect = node.getBoundingClientRect();
		var x = _Browser_window.pageXOffset;
		var y = _Browser_window.pageYOffset;
		return {
			scene: _Browser_getScene(),
			viewport: {
				x: x,
				y: y,
				width: _Browser_doc.documentElement.clientWidth,
				height: _Browser_doc.documentElement.clientHeight
			},
			element: {
				x: x + rect.left,
				y: y + rect.top,
				width: rect.width,
				height: rect.height
			}
		};
	});
}



// LOAD and RELOAD


function _Browser_reload(skipCache)
{
	return A2(elm$core$Task$perform, elm$core$Basics$never, _Scheduler_binding(function(callback)
	{
		_VirtualDom_doc.location.reload(skipCache);
	}));
}

function _Browser_load(url)
{
	return A2(elm$core$Task$perform, elm$core$Basics$never, _Scheduler_binding(function(callback)
	{
		try
		{
			_Browser_window.location = url;
		}
		catch(err)
		{
			// Only Firefox can throw a NS_ERROR_MALFORMED_URI exception here.
			// Other browsers reload the page, so let's be consistent about that.
			_VirtualDom_doc.location.reload(false);
		}
	}));
}


function _Url_percentEncode(string)
{
	return encodeURIComponent(string);
}

function _Url_percentDecode(string)
{
	try
	{
		return elm$core$Maybe$Just(decodeURIComponent(string));
	}
	catch (e)
	{
		return elm$core$Maybe$Nothing;
	}
}

// CREATE

var _Regex_never = /.^/;

var _Regex_fromStringWith = F2(function(options, string)
{
	var flags = 'g';
	if (options.multiline) { flags += 'm'; }
	if (options.caseInsensitive) { flags += 'i'; }

	try
	{
		return elm$core$Maybe$Just(new RegExp(string, flags));
	}
	catch(error)
	{
		return elm$core$Maybe$Nothing;
	}
});


// USE

var _Regex_contains = F2(function(re, string)
{
	return string.match(re) !== null;
});


var _Regex_findAtMost = F3(function(n, re, str)
{
	var out = [];
	var number = 0;
	var string = str;
	var lastIndex = re.lastIndex;
	var prevLastIndex = -1;
	var result;
	while (number++ < n && (result = re.exec(string)))
	{
		if (prevLastIndex == re.lastIndex) break;
		var i = result.length - 1;
		var subs = new Array(i);
		while (i > 0)
		{
			var submatch = result[i];
			subs[--i] = submatch
				? elm$core$Maybe$Just(submatch)
				: elm$core$Maybe$Nothing;
		}
		out.push(A4(elm$regex$Regex$Match, result[0], result.index, number, _List_fromArray(subs)));
		prevLastIndex = re.lastIndex;
	}
	re.lastIndex = lastIndex;
	return _List_fromArray(out);
});


var _Regex_replaceAtMost = F4(function(n, re, replacer, string)
{
	var count = 0;
	function jsReplacer(match)
	{
		if (count++ >= n)
		{
			return match;
		}
		var i = arguments.length - 3;
		var submatches = new Array(i);
		while (i > 0)
		{
			var submatch = arguments[i];
			submatches[--i] = submatch
				? elm$core$Maybe$Just(submatch)
				: elm$core$Maybe$Nothing;
		}
		return replacer(A4(elm$regex$Regex$Match, match, arguments[arguments.length - 2], count, _List_fromArray(submatches)));
	}
	return string.replace(re, jsReplacer);
});

var _Regex_splitAtMost = F3(function(n, re, str)
{
	var string = str;
	var out = [];
	var start = re.lastIndex;
	var restoreLastIndex = re.lastIndex;
	while (n--)
	{
		var result = re.exec(string);
		if (!result) break;
		out.push(string.slice(start, result.index));
		start = re.lastIndex;
	}
	out.push(string.slice(start));
	re.lastIndex = restoreLastIndex;
	return _List_fromArray(out);
});

var _Regex_infinity = Infinity;



var _Bitwise_and = F2(function(a, b)
{
	return a & b;
});

var _Bitwise_or = F2(function(a, b)
{
	return a | b;
});

var _Bitwise_xor = F2(function(a, b)
{
	return a ^ b;
});

function _Bitwise_complement(a)
{
	return ~a;
};

var _Bitwise_shiftLeftBy = F2(function(offset, a)
{
	return a << offset;
});

var _Bitwise_shiftRightBy = F2(function(offset, a)
{
	return a >> offset;
});

var _Bitwise_shiftRightZfBy = F2(function(offset, a)
{
	return a >>> offset;
});



// DECODER

var _File_decoder = _Json_decodePrim(function(value) {
	// NOTE: checks if `File` exists in case this is run on node
	return (typeof File === 'function' && value instanceof File)
		? elm$core$Result$Ok(value)
		: _Json_expecting('a FILE', value);
});


// METADATA

function _File_name(file) { return file.name; }
function _File_mime(file) { return file.type; }
function _File_size(file) { return file.size; }

function _File_lastModified(file)
{
	return elm$time$Time$millisToPosix(file.lastModified);
}


// DOWNLOAD

var _File_downloadNode;

function _File_getDownloadNode()
{
	return _File_downloadNode || (_File_downloadNode = document.createElementNS('http://www.w3.org/1999/xhtml', 'a'));
}

var _File_download = F3(function(name, mime, content)
{
	return _Scheduler_binding(function(callback)
	{
		var blob = new Blob([content], {type: mime});

		// for IE10+
		if (navigator.msSaveOrOpenBlob)
		{
			navigator.msSaveOrOpenBlob(blob, name);
			return;
		}

		// for HTML5
		var node = _File_getDownloadNode();
		var objectUrl = URL.createObjectURL(blob);
		node.setAttribute('href', objectUrl);
		node.setAttribute('download', name);
		node.dispatchEvent(new MouseEvent('click'));
		URL.revokeObjectURL(objectUrl);
	});
});

function _File_downloadUrl(href)
{
	return _Scheduler_binding(function(callback)
	{
		var node = _File_getDownloadNode();
		node.setAttribute('href', href);
		node.setAttribute('download', '');
		node.dispatchEvent(new MouseEvent('click'));
	});
}


// UPLOAD

function _File_uploadOne(mimes)
{
	return _Scheduler_binding(function(callback)
	{
		var node = document.createElementNS('http://www.w3.org/1999/xhtml', 'input');
		node.setAttribute('type', 'file');
		node.setAttribute('accept', A2(elm$core$String$join, ',', mimes));
		node.addEventListener('change', function(event)
		{
			callback(_Scheduler_succeed(event.target.files[0]));
		});
		node.dispatchEvent(new MouseEvent('click'));
	});
}

function _File_uploadOneOrMore(mimes)
{
	return _Scheduler_binding(function(callback)
	{
		var node = document.createElementNS('http://www.w3.org/1999/xhtml', 'input');
		node.setAttribute('type', 'file');
		node.setAttribute('accept', A2(elm$core$String$join, ',', mimes));
		node.setAttribute('multiple', '');
		node.addEventListener('change', function(event)
		{
			var elmFiles = _List_fromArray(event.target.files);
			callback(_Scheduler_succeed(_Utils_Tuple2(elmFiles.a, elmFiles.b)));
		});
		node.dispatchEvent(new MouseEvent('click'));
	});
}


// CONTENT

function _File_toString(blob)
{
	return _Scheduler_binding(function(callback)
	{
		var reader = new FileReader();
		reader.addEventListener('loadend', function() {
			callback(_Scheduler_succeed(reader.result));
		});
		reader.readAsText(blob);
		return function() { reader.abort(); };
	});
}

function _File_toBytes(blob)
{
	return _Scheduler_binding(function(callback)
	{
		var reader = new FileReader();
		reader.addEventListener('loadend', function() {
			callback(_Scheduler_succeed(new DataView(reader.result)));
		});
		reader.readAsArrayBuffer(blob);
		return function() { reader.abort(); };
	});
}

function _File_toUrl(blob)
{
	return _Scheduler_binding(function(callback)
	{
		var reader = new FileReader();
		reader.addEventListener('loadend', function() {
			callback(_Scheduler_succeed(reader.result));
		});
		reader.readAsDataURL(blob);
		return function() { reader.abort(); };
	});
}

var author$project$Main$LinkClicked = function (a) {
	return {$: 'LinkClicked', a: a};
};
var author$project$Main$UrlChanged = function (a) {
	return {$: 'UrlChanged', a: a};
};
var author$project$Commands$WhoamiResponse = F5(
	function (username, isLoggedIn, isAnon, anonIsAllowed, rights) {
		return {anonIsAllowed: anonIsAllowed, isAnon: isAnon, isLoggedIn: isLoggedIn, rights: rights, username: username};
	});
var elm$core$Array$branchFactor = 32;
var elm$core$Array$Array_elm_builtin = F4(
	function (a, b, c, d) {
		return {$: 'Array_elm_builtin', a: a, b: b, c: c, d: d};
	});
var elm$core$Basics$EQ = {$: 'EQ'};
var elm$core$Basics$GT = {$: 'GT'};
var elm$core$Basics$LT = {$: 'LT'};
var elm$core$Dict$foldr = F3(
	function (func, acc, t) {
		foldr:
		while (true) {
			if (t.$ === 'RBEmpty_elm_builtin') {
				return acc;
			} else {
				var key = t.b;
				var value = t.c;
				var left = t.d;
				var right = t.e;
				var $temp$func = func,
					$temp$acc = A3(
					func,
					key,
					value,
					A3(elm$core$Dict$foldr, func, acc, right)),
					$temp$t = left;
				func = $temp$func;
				acc = $temp$acc;
				t = $temp$t;
				continue foldr;
			}
		}
	});
var elm$core$List$cons = _List_cons;
var elm$core$Dict$toList = function (dict) {
	return A3(
		elm$core$Dict$foldr,
		F3(
			function (key, value, list) {
				return A2(
					elm$core$List$cons,
					_Utils_Tuple2(key, value),
					list);
			}),
		_List_Nil,
		dict);
};
var elm$core$Dict$keys = function (dict) {
	return A3(
		elm$core$Dict$foldr,
		F3(
			function (key, value, keyList) {
				return A2(elm$core$List$cons, key, keyList);
			}),
		_List_Nil,
		dict);
};
var elm$core$Set$toList = function (_n0) {
	var dict = _n0.a;
	return elm$core$Dict$keys(dict);
};
var elm$core$Elm$JsArray$foldr = _JsArray_foldr;
var elm$core$Array$foldr = F3(
	function (func, baseCase, _n0) {
		var tree = _n0.c;
		var tail = _n0.d;
		var helper = F2(
			function (node, acc) {
				if (node.$ === 'SubTree') {
					var subTree = node.a;
					return A3(elm$core$Elm$JsArray$foldr, helper, acc, subTree);
				} else {
					var values = node.a;
					return A3(elm$core$Elm$JsArray$foldr, func, acc, values);
				}
			});
		return A3(
			elm$core$Elm$JsArray$foldr,
			helper,
			A3(elm$core$Elm$JsArray$foldr, func, baseCase, tail),
			tree);
	});
var elm$core$Array$toList = function (array) {
	return A3(elm$core$Array$foldr, elm$core$List$cons, _List_Nil, array);
};
var elm$core$Basics$ceiling = _Basics_ceiling;
var elm$core$Basics$fdiv = _Basics_fdiv;
var elm$core$Basics$logBase = F2(
	function (base, number) {
		return _Basics_log(number) / _Basics_log(base);
	});
var elm$core$Basics$toFloat = _Basics_toFloat;
var elm$core$Array$shiftStep = elm$core$Basics$ceiling(
	A2(elm$core$Basics$logBase, 2, elm$core$Array$branchFactor));
var elm$core$Elm$JsArray$empty = _JsArray_empty;
var elm$core$Array$empty = A4(elm$core$Array$Array_elm_builtin, 0, elm$core$Array$shiftStep, elm$core$Elm$JsArray$empty, elm$core$Elm$JsArray$empty);
var elm$core$Array$Leaf = function (a) {
	return {$: 'Leaf', a: a};
};
var elm$core$Array$SubTree = function (a) {
	return {$: 'SubTree', a: a};
};
var elm$core$Elm$JsArray$initializeFromList = _JsArray_initializeFromList;
var elm$core$List$foldl = F3(
	function (func, acc, list) {
		foldl:
		while (true) {
			if (!list.b) {
				return acc;
			} else {
				var x = list.a;
				var xs = list.b;
				var $temp$func = func,
					$temp$acc = A2(func, x, acc),
					$temp$list = xs;
				func = $temp$func;
				acc = $temp$acc;
				list = $temp$list;
				continue foldl;
			}
		}
	});
var elm$core$List$reverse = function (list) {
	return A3(elm$core$List$foldl, elm$core$List$cons, _List_Nil, list);
};
var elm$core$Array$compressNodes = F2(
	function (nodes, acc) {
		compressNodes:
		while (true) {
			var _n0 = A2(elm$core$Elm$JsArray$initializeFromList, elm$core$Array$branchFactor, nodes);
			var node = _n0.a;
			var remainingNodes = _n0.b;
			var newAcc = A2(
				elm$core$List$cons,
				elm$core$Array$SubTree(node),
				acc);
			if (!remainingNodes.b) {
				return elm$core$List$reverse(newAcc);
			} else {
				var $temp$nodes = remainingNodes,
					$temp$acc = newAcc;
				nodes = $temp$nodes;
				acc = $temp$acc;
				continue compressNodes;
			}
		}
	});
var elm$core$Basics$apR = F2(
	function (x, f) {
		return f(x);
	});
var elm$core$Basics$eq = _Utils_equal;
var elm$core$Tuple$first = function (_n0) {
	var x = _n0.a;
	return x;
};
var elm$core$Array$treeFromBuilder = F2(
	function (nodeList, nodeListSize) {
		treeFromBuilder:
		while (true) {
			var newNodeSize = elm$core$Basics$ceiling(nodeListSize / elm$core$Array$branchFactor);
			if (newNodeSize === 1) {
				return A2(elm$core$Elm$JsArray$initializeFromList, elm$core$Array$branchFactor, nodeList).a;
			} else {
				var $temp$nodeList = A2(elm$core$Array$compressNodes, nodeList, _List_Nil),
					$temp$nodeListSize = newNodeSize;
				nodeList = $temp$nodeList;
				nodeListSize = $temp$nodeListSize;
				continue treeFromBuilder;
			}
		}
	});
var elm$core$Basics$add = _Basics_add;
var elm$core$Basics$apL = F2(
	function (f, x) {
		return f(x);
	});
var elm$core$Basics$floor = _Basics_floor;
var elm$core$Basics$gt = _Utils_gt;
var elm$core$Basics$max = F2(
	function (x, y) {
		return (_Utils_cmp(x, y) > 0) ? x : y;
	});
var elm$core$Basics$mul = _Basics_mul;
var elm$core$Basics$sub = _Basics_sub;
var elm$core$Elm$JsArray$length = _JsArray_length;
var elm$core$Array$builderToArray = F2(
	function (reverseNodeList, builder) {
		if (!builder.nodeListSize) {
			return A4(
				elm$core$Array$Array_elm_builtin,
				elm$core$Elm$JsArray$length(builder.tail),
				elm$core$Array$shiftStep,
				elm$core$Elm$JsArray$empty,
				builder.tail);
		} else {
			var treeLen = builder.nodeListSize * elm$core$Array$branchFactor;
			var depth = elm$core$Basics$floor(
				A2(elm$core$Basics$logBase, elm$core$Array$branchFactor, treeLen - 1));
			var correctNodeList = reverseNodeList ? elm$core$List$reverse(builder.nodeList) : builder.nodeList;
			var tree = A2(elm$core$Array$treeFromBuilder, correctNodeList, builder.nodeListSize);
			return A4(
				elm$core$Array$Array_elm_builtin,
				elm$core$Elm$JsArray$length(builder.tail) + treeLen,
				A2(elm$core$Basics$max, 5, depth * elm$core$Array$shiftStep),
				tree,
				builder.tail);
		}
	});
var elm$core$Basics$False = {$: 'False'};
var elm$core$Basics$idiv = _Basics_idiv;
var elm$core$Basics$lt = _Utils_lt;
var elm$core$Elm$JsArray$initialize = _JsArray_initialize;
var elm$core$Array$initializeHelp = F5(
	function (fn, fromIndex, len, nodeList, tail) {
		initializeHelp:
		while (true) {
			if (fromIndex < 0) {
				return A2(
					elm$core$Array$builderToArray,
					false,
					{nodeList: nodeList, nodeListSize: (len / elm$core$Array$branchFactor) | 0, tail: tail});
			} else {
				var leaf = elm$core$Array$Leaf(
					A3(elm$core$Elm$JsArray$initialize, elm$core$Array$branchFactor, fromIndex, fn));
				var $temp$fn = fn,
					$temp$fromIndex = fromIndex - elm$core$Array$branchFactor,
					$temp$len = len,
					$temp$nodeList = A2(elm$core$List$cons, leaf, nodeList),
					$temp$tail = tail;
				fn = $temp$fn;
				fromIndex = $temp$fromIndex;
				len = $temp$len;
				nodeList = $temp$nodeList;
				tail = $temp$tail;
				continue initializeHelp;
			}
		}
	});
var elm$core$Basics$le = _Utils_le;
var elm$core$Basics$remainderBy = _Basics_remainderBy;
var elm$core$Array$initialize = F2(
	function (len, fn) {
		if (len <= 0) {
			return elm$core$Array$empty;
		} else {
			var tailLen = len % elm$core$Array$branchFactor;
			var tail = A3(elm$core$Elm$JsArray$initialize, tailLen, len - tailLen, fn);
			var initialFromIndex = (len - tailLen) - elm$core$Array$branchFactor;
			return A5(elm$core$Array$initializeHelp, fn, initialFromIndex, len, _List_Nil, tail);
		}
	});
var elm$core$Maybe$Just = function (a) {
	return {$: 'Just', a: a};
};
var elm$core$Maybe$Nothing = {$: 'Nothing'};
var elm$core$Result$Err = function (a) {
	return {$: 'Err', a: a};
};
var elm$core$Result$Ok = function (a) {
	return {$: 'Ok', a: a};
};
var elm$core$Basics$True = {$: 'True'};
var elm$core$Result$isOk = function (result) {
	if (result.$ === 'Ok') {
		return true;
	} else {
		return false;
	}
};
var elm$json$Json$Decode$Failure = F2(
	function (a, b) {
		return {$: 'Failure', a: a, b: b};
	});
var elm$json$Json$Decode$Field = F2(
	function (a, b) {
		return {$: 'Field', a: a, b: b};
	});
var elm$json$Json$Decode$Index = F2(
	function (a, b) {
		return {$: 'Index', a: a, b: b};
	});
var elm$json$Json$Decode$OneOf = function (a) {
	return {$: 'OneOf', a: a};
};
var elm$core$Basics$and = _Basics_and;
var elm$core$Basics$append = _Utils_append;
var elm$core$Basics$or = _Basics_or;
var elm$core$Char$toCode = _Char_toCode;
var elm$core$Char$isLower = function (_char) {
	var code = elm$core$Char$toCode(_char);
	return (97 <= code) && (code <= 122);
};
var elm$core$Char$isUpper = function (_char) {
	var code = elm$core$Char$toCode(_char);
	return (code <= 90) && (65 <= code);
};
var elm$core$Char$isAlpha = function (_char) {
	return elm$core$Char$isLower(_char) || elm$core$Char$isUpper(_char);
};
var elm$core$Char$isDigit = function (_char) {
	var code = elm$core$Char$toCode(_char);
	return (code <= 57) && (48 <= code);
};
var elm$core$Char$isAlphaNum = function (_char) {
	return elm$core$Char$isLower(_char) || (elm$core$Char$isUpper(_char) || elm$core$Char$isDigit(_char));
};
var elm$core$List$length = function (xs) {
	return A3(
		elm$core$List$foldl,
		F2(
			function (_n0, i) {
				return i + 1;
			}),
		0,
		xs);
};
var elm$core$List$map2 = _List_map2;
var elm$core$List$rangeHelp = F3(
	function (lo, hi, list) {
		rangeHelp:
		while (true) {
			if (_Utils_cmp(lo, hi) < 1) {
				var $temp$lo = lo,
					$temp$hi = hi - 1,
					$temp$list = A2(elm$core$List$cons, hi, list);
				lo = $temp$lo;
				hi = $temp$hi;
				list = $temp$list;
				continue rangeHelp;
			} else {
				return list;
			}
		}
	});
var elm$core$List$range = F2(
	function (lo, hi) {
		return A3(elm$core$List$rangeHelp, lo, hi, _List_Nil);
	});
var elm$core$List$indexedMap = F2(
	function (f, xs) {
		return A3(
			elm$core$List$map2,
			f,
			A2(
				elm$core$List$range,
				0,
				elm$core$List$length(xs) - 1),
			xs);
	});
var elm$core$String$all = _String_all;
var elm$core$String$fromInt = _String_fromNumber;
var elm$core$String$join = F2(
	function (sep, chunks) {
		return A2(
			_String_join,
			sep,
			_List_toArray(chunks));
	});
var elm$core$String$uncons = _String_uncons;
var elm$core$String$split = F2(
	function (sep, string) {
		return _List_fromArray(
			A2(_String_split, sep, string));
	});
var elm$json$Json$Decode$indent = function (str) {
	return A2(
		elm$core$String$join,
		'\n    ',
		A2(elm$core$String$split, '\n', str));
};
var elm$json$Json$Encode$encode = _Json_encode;
var elm$json$Json$Decode$errorOneOf = F2(
	function (i, error) {
		return '\n\n(' + (elm$core$String$fromInt(i + 1) + (') ' + elm$json$Json$Decode$indent(
			elm$json$Json$Decode$errorToString(error))));
	});
var elm$json$Json$Decode$errorToString = function (error) {
	return A2(elm$json$Json$Decode$errorToStringHelp, error, _List_Nil);
};
var elm$json$Json$Decode$errorToStringHelp = F2(
	function (error, context) {
		errorToStringHelp:
		while (true) {
			switch (error.$) {
				case 'Field':
					var f = error.a;
					var err = error.b;
					var isSimple = function () {
						var _n1 = elm$core$String$uncons(f);
						if (_n1.$ === 'Nothing') {
							return false;
						} else {
							var _n2 = _n1.a;
							var _char = _n2.a;
							var rest = _n2.b;
							return elm$core$Char$isAlpha(_char) && A2(elm$core$String$all, elm$core$Char$isAlphaNum, rest);
						}
					}();
					var fieldName = isSimple ? ('.' + f) : ('[\'' + (f + '\']'));
					var $temp$error = err,
						$temp$context = A2(elm$core$List$cons, fieldName, context);
					error = $temp$error;
					context = $temp$context;
					continue errorToStringHelp;
				case 'Index':
					var i = error.a;
					var err = error.b;
					var indexName = '[' + (elm$core$String$fromInt(i) + ']');
					var $temp$error = err,
						$temp$context = A2(elm$core$List$cons, indexName, context);
					error = $temp$error;
					context = $temp$context;
					continue errorToStringHelp;
				case 'OneOf':
					var errors = error.a;
					if (!errors.b) {
						return 'Ran into a Json.Decode.oneOf with no possibilities' + function () {
							if (!context.b) {
								return '!';
							} else {
								return ' at json' + A2(
									elm$core$String$join,
									'',
									elm$core$List$reverse(context));
							}
						}();
					} else {
						if (!errors.b.b) {
							var err = errors.a;
							var $temp$error = err,
								$temp$context = context;
							error = $temp$error;
							context = $temp$context;
							continue errorToStringHelp;
						} else {
							var starter = function () {
								if (!context.b) {
									return 'Json.Decode.oneOf';
								} else {
									return 'The Json.Decode.oneOf at json' + A2(
										elm$core$String$join,
										'',
										elm$core$List$reverse(context));
								}
							}();
							var introduction = starter + (' failed in the following ' + (elm$core$String$fromInt(
								elm$core$List$length(errors)) + ' ways:'));
							return A2(
								elm$core$String$join,
								'\n\n',
								A2(
									elm$core$List$cons,
									introduction,
									A2(elm$core$List$indexedMap, elm$json$Json$Decode$errorOneOf, errors)));
						}
					}
				default:
					var msg = error.a;
					var json = error.b;
					var introduction = function () {
						if (!context.b) {
							return 'Problem with the given value:\n\n';
						} else {
							return 'Problem with the value at json' + (A2(
								elm$core$String$join,
								'',
								elm$core$List$reverse(context)) + ':\n\n    ');
						}
					}();
					return introduction + (elm$json$Json$Decode$indent(
						A2(elm$json$Json$Encode$encode, 4, json)) + ('\n\n' + msg));
			}
		}
	});
var elm$json$Json$Decode$bool = _Json_decodeBool;
var elm$json$Json$Decode$field = _Json_decodeField;
var elm$json$Json$Decode$list = _Json_decodeList;
var elm$json$Json$Decode$map5 = _Json_map5;
var elm$json$Json$Decode$string = _Json_decodeString;
var author$project$Commands$decodeWhoami = A6(
	elm$json$Json$Decode$map5,
	author$project$Commands$WhoamiResponse,
	A2(elm$json$Json$Decode$field, 'user', elm$json$Json$Decode$string),
	A2(elm$json$Json$Decode$field, 'is_logged_in', elm$json$Json$Decode$bool),
	A2(elm$json$Json$Decode$field, 'is_anon', elm$json$Json$Decode$bool),
	A2(elm$json$Json$Decode$field, 'anon_is_allowed', elm$json$Json$Decode$bool),
	A2(
		elm$json$Json$Decode$field,
		'rights',
		elm$json$Json$Decode$list(elm$json$Json$Decode$string)));
var elm$core$Dict$RBEmpty_elm_builtin = {$: 'RBEmpty_elm_builtin'};
var elm$core$Dict$empty = elm$core$Dict$RBEmpty_elm_builtin;
var elm$core$Basics$compare = _Utils_compare;
var elm$core$Dict$get = F2(
	function (targetKey, dict) {
		get:
		while (true) {
			if (dict.$ === 'RBEmpty_elm_builtin') {
				return elm$core$Maybe$Nothing;
			} else {
				var key = dict.b;
				var value = dict.c;
				var left = dict.d;
				var right = dict.e;
				var _n1 = A2(elm$core$Basics$compare, targetKey, key);
				switch (_n1.$) {
					case 'LT':
						var $temp$targetKey = targetKey,
							$temp$dict = left;
						targetKey = $temp$targetKey;
						dict = $temp$dict;
						continue get;
					case 'EQ':
						return elm$core$Maybe$Just(value);
					default:
						var $temp$targetKey = targetKey,
							$temp$dict = right;
						targetKey = $temp$targetKey;
						dict = $temp$dict;
						continue get;
				}
			}
		}
	});
var elm$core$Dict$Black = {$: 'Black'};
var elm$core$Dict$RBNode_elm_builtin = F5(
	function (a, b, c, d, e) {
		return {$: 'RBNode_elm_builtin', a: a, b: b, c: c, d: d, e: e};
	});
var elm$core$Dict$Red = {$: 'Red'};
var elm$core$Dict$balance = F5(
	function (color, key, value, left, right) {
		if ((right.$ === 'RBNode_elm_builtin') && (right.a.$ === 'Red')) {
			var _n1 = right.a;
			var rK = right.b;
			var rV = right.c;
			var rLeft = right.d;
			var rRight = right.e;
			if ((left.$ === 'RBNode_elm_builtin') && (left.a.$ === 'Red')) {
				var _n3 = left.a;
				var lK = left.b;
				var lV = left.c;
				var lLeft = left.d;
				var lRight = left.e;
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Red,
					key,
					value,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, lK, lV, lLeft, lRight),
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, rK, rV, rLeft, rRight));
			} else {
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					color,
					rK,
					rV,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, key, value, left, rLeft),
					rRight);
			}
		} else {
			if ((((left.$ === 'RBNode_elm_builtin') && (left.a.$ === 'Red')) && (left.d.$ === 'RBNode_elm_builtin')) && (left.d.a.$ === 'Red')) {
				var _n5 = left.a;
				var lK = left.b;
				var lV = left.c;
				var _n6 = left.d;
				var _n7 = _n6.a;
				var llK = _n6.b;
				var llV = _n6.c;
				var llLeft = _n6.d;
				var llRight = _n6.e;
				var lRight = left.e;
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Red,
					lK,
					lV,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, llK, llV, llLeft, llRight),
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, key, value, lRight, right));
			} else {
				return A5(elm$core$Dict$RBNode_elm_builtin, color, key, value, left, right);
			}
		}
	});
var elm$core$Dict$insertHelp = F3(
	function (key, value, dict) {
		if (dict.$ === 'RBEmpty_elm_builtin') {
			return A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, key, value, elm$core$Dict$RBEmpty_elm_builtin, elm$core$Dict$RBEmpty_elm_builtin);
		} else {
			var nColor = dict.a;
			var nKey = dict.b;
			var nValue = dict.c;
			var nLeft = dict.d;
			var nRight = dict.e;
			var _n1 = A2(elm$core$Basics$compare, key, nKey);
			switch (_n1.$) {
				case 'LT':
					return A5(
						elm$core$Dict$balance,
						nColor,
						nKey,
						nValue,
						A3(elm$core$Dict$insertHelp, key, value, nLeft),
						nRight);
				case 'EQ':
					return A5(elm$core$Dict$RBNode_elm_builtin, nColor, nKey, value, nLeft, nRight);
				default:
					return A5(
						elm$core$Dict$balance,
						nColor,
						nKey,
						nValue,
						nLeft,
						A3(elm$core$Dict$insertHelp, key, value, nRight));
			}
		}
	});
var elm$core$Dict$insert = F3(
	function (key, value, dict) {
		var _n0 = A3(elm$core$Dict$insertHelp, key, value, dict);
		if ((_n0.$ === 'RBNode_elm_builtin') && (_n0.a.$ === 'Red')) {
			var _n1 = _n0.a;
			var k = _n0.b;
			var v = _n0.c;
			var l = _n0.d;
			var r = _n0.e;
			return A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, k, v, l, r);
		} else {
			var x = _n0;
			return x;
		}
	});
var elm$core$Dict$getMin = function (dict) {
	getMin:
	while (true) {
		if ((dict.$ === 'RBNode_elm_builtin') && (dict.d.$ === 'RBNode_elm_builtin')) {
			var left = dict.d;
			var $temp$dict = left;
			dict = $temp$dict;
			continue getMin;
		} else {
			return dict;
		}
	}
};
var elm$core$Dict$moveRedLeft = function (dict) {
	if (((dict.$ === 'RBNode_elm_builtin') && (dict.d.$ === 'RBNode_elm_builtin')) && (dict.e.$ === 'RBNode_elm_builtin')) {
		if ((dict.e.d.$ === 'RBNode_elm_builtin') && (dict.e.d.a.$ === 'Red')) {
			var clr = dict.a;
			var k = dict.b;
			var v = dict.c;
			var _n1 = dict.d;
			var lClr = _n1.a;
			var lK = _n1.b;
			var lV = _n1.c;
			var lLeft = _n1.d;
			var lRight = _n1.e;
			var _n2 = dict.e;
			var rClr = _n2.a;
			var rK = _n2.b;
			var rV = _n2.c;
			var rLeft = _n2.d;
			var _n3 = rLeft.a;
			var rlK = rLeft.b;
			var rlV = rLeft.c;
			var rlL = rLeft.d;
			var rlR = rLeft.e;
			var rRight = _n2.e;
			return A5(
				elm$core$Dict$RBNode_elm_builtin,
				elm$core$Dict$Red,
				rlK,
				rlV,
				A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Black,
					k,
					v,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, lK, lV, lLeft, lRight),
					rlL),
				A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, rK, rV, rlR, rRight));
		} else {
			var clr = dict.a;
			var k = dict.b;
			var v = dict.c;
			var _n4 = dict.d;
			var lClr = _n4.a;
			var lK = _n4.b;
			var lV = _n4.c;
			var lLeft = _n4.d;
			var lRight = _n4.e;
			var _n5 = dict.e;
			var rClr = _n5.a;
			var rK = _n5.b;
			var rV = _n5.c;
			var rLeft = _n5.d;
			var rRight = _n5.e;
			if (clr.$ === 'Black') {
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Black,
					k,
					v,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, lK, lV, lLeft, lRight),
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, rK, rV, rLeft, rRight));
			} else {
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Black,
					k,
					v,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, lK, lV, lLeft, lRight),
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, rK, rV, rLeft, rRight));
			}
		}
	} else {
		return dict;
	}
};
var elm$core$Dict$moveRedRight = function (dict) {
	if (((dict.$ === 'RBNode_elm_builtin') && (dict.d.$ === 'RBNode_elm_builtin')) && (dict.e.$ === 'RBNode_elm_builtin')) {
		if ((dict.d.d.$ === 'RBNode_elm_builtin') && (dict.d.d.a.$ === 'Red')) {
			var clr = dict.a;
			var k = dict.b;
			var v = dict.c;
			var _n1 = dict.d;
			var lClr = _n1.a;
			var lK = _n1.b;
			var lV = _n1.c;
			var _n2 = _n1.d;
			var _n3 = _n2.a;
			var llK = _n2.b;
			var llV = _n2.c;
			var llLeft = _n2.d;
			var llRight = _n2.e;
			var lRight = _n1.e;
			var _n4 = dict.e;
			var rClr = _n4.a;
			var rK = _n4.b;
			var rV = _n4.c;
			var rLeft = _n4.d;
			var rRight = _n4.e;
			return A5(
				elm$core$Dict$RBNode_elm_builtin,
				elm$core$Dict$Red,
				lK,
				lV,
				A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, llK, llV, llLeft, llRight),
				A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Black,
					k,
					v,
					lRight,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, rK, rV, rLeft, rRight)));
		} else {
			var clr = dict.a;
			var k = dict.b;
			var v = dict.c;
			var _n5 = dict.d;
			var lClr = _n5.a;
			var lK = _n5.b;
			var lV = _n5.c;
			var lLeft = _n5.d;
			var lRight = _n5.e;
			var _n6 = dict.e;
			var rClr = _n6.a;
			var rK = _n6.b;
			var rV = _n6.c;
			var rLeft = _n6.d;
			var rRight = _n6.e;
			if (clr.$ === 'Black') {
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Black,
					k,
					v,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, lK, lV, lLeft, lRight),
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, rK, rV, rLeft, rRight));
			} else {
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					elm$core$Dict$Black,
					k,
					v,
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, lK, lV, lLeft, lRight),
					A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, rK, rV, rLeft, rRight));
			}
		}
	} else {
		return dict;
	}
};
var elm$core$Dict$removeHelpPrepEQGT = F7(
	function (targetKey, dict, color, key, value, left, right) {
		if ((left.$ === 'RBNode_elm_builtin') && (left.a.$ === 'Red')) {
			var _n1 = left.a;
			var lK = left.b;
			var lV = left.c;
			var lLeft = left.d;
			var lRight = left.e;
			return A5(
				elm$core$Dict$RBNode_elm_builtin,
				color,
				lK,
				lV,
				lLeft,
				A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Red, key, value, lRight, right));
		} else {
			_n2$2:
			while (true) {
				if ((right.$ === 'RBNode_elm_builtin') && (right.a.$ === 'Black')) {
					if (right.d.$ === 'RBNode_elm_builtin') {
						if (right.d.a.$ === 'Black') {
							var _n3 = right.a;
							var _n4 = right.d;
							var _n5 = _n4.a;
							return elm$core$Dict$moveRedRight(dict);
						} else {
							break _n2$2;
						}
					} else {
						var _n6 = right.a;
						var _n7 = right.d;
						return elm$core$Dict$moveRedRight(dict);
					}
				} else {
					break _n2$2;
				}
			}
			return dict;
		}
	});
var elm$core$Dict$removeMin = function (dict) {
	if ((dict.$ === 'RBNode_elm_builtin') && (dict.d.$ === 'RBNode_elm_builtin')) {
		var color = dict.a;
		var key = dict.b;
		var value = dict.c;
		var left = dict.d;
		var lColor = left.a;
		var lLeft = left.d;
		var right = dict.e;
		if (lColor.$ === 'Black') {
			if ((lLeft.$ === 'RBNode_elm_builtin') && (lLeft.a.$ === 'Red')) {
				var _n3 = lLeft.a;
				return A5(
					elm$core$Dict$RBNode_elm_builtin,
					color,
					key,
					value,
					elm$core$Dict$removeMin(left),
					right);
			} else {
				var _n4 = elm$core$Dict$moveRedLeft(dict);
				if (_n4.$ === 'RBNode_elm_builtin') {
					var nColor = _n4.a;
					var nKey = _n4.b;
					var nValue = _n4.c;
					var nLeft = _n4.d;
					var nRight = _n4.e;
					return A5(
						elm$core$Dict$balance,
						nColor,
						nKey,
						nValue,
						elm$core$Dict$removeMin(nLeft),
						nRight);
				} else {
					return elm$core$Dict$RBEmpty_elm_builtin;
				}
			}
		} else {
			return A5(
				elm$core$Dict$RBNode_elm_builtin,
				color,
				key,
				value,
				elm$core$Dict$removeMin(left),
				right);
		}
	} else {
		return elm$core$Dict$RBEmpty_elm_builtin;
	}
};
var elm$core$Dict$removeHelp = F2(
	function (targetKey, dict) {
		if (dict.$ === 'RBEmpty_elm_builtin') {
			return elm$core$Dict$RBEmpty_elm_builtin;
		} else {
			var color = dict.a;
			var key = dict.b;
			var value = dict.c;
			var left = dict.d;
			var right = dict.e;
			if (_Utils_cmp(targetKey, key) < 0) {
				if ((left.$ === 'RBNode_elm_builtin') && (left.a.$ === 'Black')) {
					var _n4 = left.a;
					var lLeft = left.d;
					if ((lLeft.$ === 'RBNode_elm_builtin') && (lLeft.a.$ === 'Red')) {
						var _n6 = lLeft.a;
						return A5(
							elm$core$Dict$RBNode_elm_builtin,
							color,
							key,
							value,
							A2(elm$core$Dict$removeHelp, targetKey, left),
							right);
					} else {
						var _n7 = elm$core$Dict$moveRedLeft(dict);
						if (_n7.$ === 'RBNode_elm_builtin') {
							var nColor = _n7.a;
							var nKey = _n7.b;
							var nValue = _n7.c;
							var nLeft = _n7.d;
							var nRight = _n7.e;
							return A5(
								elm$core$Dict$balance,
								nColor,
								nKey,
								nValue,
								A2(elm$core$Dict$removeHelp, targetKey, nLeft),
								nRight);
						} else {
							return elm$core$Dict$RBEmpty_elm_builtin;
						}
					}
				} else {
					return A5(
						elm$core$Dict$RBNode_elm_builtin,
						color,
						key,
						value,
						A2(elm$core$Dict$removeHelp, targetKey, left),
						right);
				}
			} else {
				return A2(
					elm$core$Dict$removeHelpEQGT,
					targetKey,
					A7(elm$core$Dict$removeHelpPrepEQGT, targetKey, dict, color, key, value, left, right));
			}
		}
	});
var elm$core$Dict$removeHelpEQGT = F2(
	function (targetKey, dict) {
		if (dict.$ === 'RBNode_elm_builtin') {
			var color = dict.a;
			var key = dict.b;
			var value = dict.c;
			var left = dict.d;
			var right = dict.e;
			if (_Utils_eq(targetKey, key)) {
				var _n1 = elm$core$Dict$getMin(right);
				if (_n1.$ === 'RBNode_elm_builtin') {
					var minKey = _n1.b;
					var minValue = _n1.c;
					return A5(
						elm$core$Dict$balance,
						color,
						minKey,
						minValue,
						left,
						elm$core$Dict$removeMin(right));
				} else {
					return elm$core$Dict$RBEmpty_elm_builtin;
				}
			} else {
				return A5(
					elm$core$Dict$balance,
					color,
					key,
					value,
					left,
					A2(elm$core$Dict$removeHelp, targetKey, right));
			}
		} else {
			return elm$core$Dict$RBEmpty_elm_builtin;
		}
	});
var elm$core$Dict$remove = F2(
	function (key, dict) {
		var _n0 = A2(elm$core$Dict$removeHelp, key, dict);
		if ((_n0.$ === 'RBNode_elm_builtin') && (_n0.a.$ === 'Red')) {
			var _n1 = _n0.a;
			var k = _n0.b;
			var v = _n0.c;
			var l = _n0.d;
			var r = _n0.e;
			return A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, k, v, l, r);
		} else {
			var x = _n0;
			return x;
		}
	});
var elm$core$Dict$update = F3(
	function (targetKey, alter, dictionary) {
		var _n0 = alter(
			A2(elm$core$Dict$get, targetKey, dictionary));
		if (_n0.$ === 'Just') {
			var value = _n0.a;
			return A3(elm$core$Dict$insert, targetKey, value, dictionary);
		} else {
			return A2(elm$core$Dict$remove, targetKey, dictionary);
		}
	});
var elm$core$Maybe$isJust = function (maybe) {
	if (maybe.$ === 'Just') {
		return true;
	} else {
		return false;
	}
};
var elm$core$Platform$sendToApp = _Platform_sendToApp;
var elm$core$Platform$sendToSelf = _Platform_sendToSelf;
var elm$core$Result$map = F2(
	function (func, ra) {
		if (ra.$ === 'Ok') {
			var a = ra.a;
			return elm$core$Result$Ok(
				func(a));
		} else {
			var e = ra.a;
			return elm$core$Result$Err(e);
		}
	});
var elm$http$Http$BadStatus_ = F2(
	function (a, b) {
		return {$: 'BadStatus_', a: a, b: b};
	});
var elm$http$Http$BadUrl_ = function (a) {
	return {$: 'BadUrl_', a: a};
};
var elm$http$Http$GoodStatus_ = F2(
	function (a, b) {
		return {$: 'GoodStatus_', a: a, b: b};
	});
var elm$http$Http$NetworkError_ = {$: 'NetworkError_'};
var elm$http$Http$Receiving = function (a) {
	return {$: 'Receiving', a: a};
};
var elm$http$Http$Sending = function (a) {
	return {$: 'Sending', a: a};
};
var elm$http$Http$Timeout_ = {$: 'Timeout_'};
var elm$http$Http$emptyBody = _Http_emptyBody;
var elm$core$Result$mapError = F2(
	function (f, result) {
		if (result.$ === 'Ok') {
			var v = result.a;
			return elm$core$Result$Ok(v);
		} else {
			var e = result.a;
			return elm$core$Result$Err(
				f(e));
		}
	});
var elm$core$Basics$composeR = F3(
	function (f, g, x) {
		return g(
			f(x));
	});
var elm$core$Basics$identity = function (x) {
	return x;
};
var elm$http$Http$expectStringResponse = F2(
	function (toMsg, toResult) {
		return A3(
			_Http_expect,
			'',
			elm$core$Basics$identity,
			A2(elm$core$Basics$composeR, toResult, toMsg));
	});
var elm$http$Http$BadBody = function (a) {
	return {$: 'BadBody', a: a};
};
var elm$http$Http$BadStatus = function (a) {
	return {$: 'BadStatus', a: a};
};
var elm$http$Http$BadUrl = function (a) {
	return {$: 'BadUrl', a: a};
};
var elm$http$Http$NetworkError = {$: 'NetworkError'};
var elm$http$Http$Timeout = {$: 'Timeout'};
var elm$http$Http$resolve = F2(
	function (toResult, response) {
		switch (response.$) {
			case 'BadUrl_':
				var url = response.a;
				return elm$core$Result$Err(
					elm$http$Http$BadUrl(url));
			case 'Timeout_':
				return elm$core$Result$Err(elm$http$Http$Timeout);
			case 'NetworkError_':
				return elm$core$Result$Err(elm$http$Http$NetworkError);
			case 'BadStatus_':
				var metadata = response.a;
				return elm$core$Result$Err(
					elm$http$Http$BadStatus(metadata.statusCode));
			default:
				var body = response.b;
				return A2(
					elm$core$Result$mapError,
					elm$http$Http$BadBody,
					toResult(body));
		}
	});
var elm$json$Json$Decode$decodeString = _Json_runOnString;
var elm$http$Http$expectJson = F2(
	function (toMsg, decoder) {
		return A2(
			elm$http$Http$expectStringResponse,
			toMsg,
			elm$http$Http$resolve(
				function (string) {
					return A2(
						elm$core$Result$mapError,
						elm$json$Json$Decode$errorToString,
						A2(elm$json$Json$Decode$decodeString, decoder, string));
				}));
	});
var elm$http$Http$Request = function (a) {
	return {$: 'Request', a: a};
};
var elm$core$Task$succeed = _Scheduler_succeed;
var elm$http$Http$State = F2(
	function (reqs, subs) {
		return {reqs: reqs, subs: subs};
	});
var elm$http$Http$init = elm$core$Task$succeed(
	A2(elm$http$Http$State, elm$core$Dict$empty, _List_Nil));
var elm$core$Task$andThen = _Scheduler_andThen;
var elm$core$Process$kill = _Scheduler_kill;
var elm$core$Process$spawn = _Scheduler_spawn;
var elm$http$Http$updateReqs = F3(
	function (router, cmds, reqs) {
		updateReqs:
		while (true) {
			if (!cmds.b) {
				return elm$core$Task$succeed(reqs);
			} else {
				var cmd = cmds.a;
				var otherCmds = cmds.b;
				if (cmd.$ === 'Cancel') {
					var tracker = cmd.a;
					var _n2 = A2(elm$core$Dict$get, tracker, reqs);
					if (_n2.$ === 'Nothing') {
						var $temp$router = router,
							$temp$cmds = otherCmds,
							$temp$reqs = reqs;
						router = $temp$router;
						cmds = $temp$cmds;
						reqs = $temp$reqs;
						continue updateReqs;
					} else {
						var pid = _n2.a;
						return A2(
							elm$core$Task$andThen,
							function (_n3) {
								return A3(
									elm$http$Http$updateReqs,
									router,
									otherCmds,
									A2(elm$core$Dict$remove, tracker, reqs));
							},
							elm$core$Process$kill(pid));
					}
				} else {
					var req = cmd.a;
					return A2(
						elm$core$Task$andThen,
						function (pid) {
							var _n4 = req.tracker;
							if (_n4.$ === 'Nothing') {
								return A3(elm$http$Http$updateReqs, router, otherCmds, reqs);
							} else {
								var tracker = _n4.a;
								return A3(
									elm$http$Http$updateReqs,
									router,
									otherCmds,
									A3(elm$core$Dict$insert, tracker, pid, reqs));
							}
						},
						elm$core$Process$spawn(
							A3(
								_Http_toTask,
								router,
								elm$core$Platform$sendToApp(router),
								req)));
				}
			}
		}
	});
var elm$http$Http$onEffects = F4(
	function (router, cmds, subs, state) {
		return A2(
			elm$core$Task$andThen,
			function (reqs) {
				return elm$core$Task$succeed(
					A2(elm$http$Http$State, reqs, subs));
			},
			A3(elm$http$Http$updateReqs, router, cmds, state.reqs));
	});
var elm$core$List$foldrHelper = F4(
	function (fn, acc, ctr, ls) {
		if (!ls.b) {
			return acc;
		} else {
			var a = ls.a;
			var r1 = ls.b;
			if (!r1.b) {
				return A2(fn, a, acc);
			} else {
				var b = r1.a;
				var r2 = r1.b;
				if (!r2.b) {
					return A2(
						fn,
						a,
						A2(fn, b, acc));
				} else {
					var c = r2.a;
					var r3 = r2.b;
					if (!r3.b) {
						return A2(
							fn,
							a,
							A2(
								fn,
								b,
								A2(fn, c, acc)));
					} else {
						var d = r3.a;
						var r4 = r3.b;
						var res = (ctr > 500) ? A3(
							elm$core$List$foldl,
							fn,
							acc,
							elm$core$List$reverse(r4)) : A4(elm$core$List$foldrHelper, fn, acc, ctr + 1, r4);
						return A2(
							fn,
							a,
							A2(
								fn,
								b,
								A2(
									fn,
									c,
									A2(fn, d, res))));
					}
				}
			}
		}
	});
var elm$core$List$foldr = F3(
	function (fn, acc, ls) {
		return A4(elm$core$List$foldrHelper, fn, acc, 0, ls);
	});
var elm$core$List$maybeCons = F3(
	function (f, mx, xs) {
		var _n0 = f(mx);
		if (_n0.$ === 'Just') {
			var x = _n0.a;
			return A2(elm$core$List$cons, x, xs);
		} else {
			return xs;
		}
	});
var elm$core$List$filterMap = F2(
	function (f, xs) {
		return A3(
			elm$core$List$foldr,
			elm$core$List$maybeCons(f),
			_List_Nil,
			xs);
	});
var elm$core$Task$map2 = F3(
	function (func, taskA, taskB) {
		return A2(
			elm$core$Task$andThen,
			function (a) {
				return A2(
					elm$core$Task$andThen,
					function (b) {
						return elm$core$Task$succeed(
							A2(func, a, b));
					},
					taskB);
			},
			taskA);
	});
var elm$core$Task$sequence = function (tasks) {
	return A3(
		elm$core$List$foldr,
		elm$core$Task$map2(elm$core$List$cons),
		elm$core$Task$succeed(_List_Nil),
		tasks);
};
var elm$http$Http$maybeSend = F4(
	function (router, desiredTracker, progress, _n0) {
		var actualTracker = _n0.a;
		var toMsg = _n0.b;
		return _Utils_eq(desiredTracker, actualTracker) ? elm$core$Maybe$Just(
			A2(
				elm$core$Platform$sendToApp,
				router,
				toMsg(progress))) : elm$core$Maybe$Nothing;
	});
var elm$http$Http$onSelfMsg = F3(
	function (router, _n0, state) {
		var tracker = _n0.a;
		var progress = _n0.b;
		return A2(
			elm$core$Task$andThen,
			function (_n1) {
				return elm$core$Task$succeed(state);
			},
			elm$core$Task$sequence(
				A2(
					elm$core$List$filterMap,
					A3(elm$http$Http$maybeSend, router, tracker, progress),
					state.subs)));
	});
var elm$http$Http$Cancel = function (a) {
	return {$: 'Cancel', a: a};
};
var elm$http$Http$cmdMap = F2(
	function (func, cmd) {
		if (cmd.$ === 'Cancel') {
			var tracker = cmd.a;
			return elm$http$Http$Cancel(tracker);
		} else {
			var r = cmd.a;
			return elm$http$Http$Request(
				{
					allowCookiesFromOtherDomains: r.allowCookiesFromOtherDomains,
					body: r.body,
					expect: A2(_Http_mapExpect, func, r.expect),
					headers: r.headers,
					method: r.method,
					timeout: r.timeout,
					tracker: r.tracker,
					url: r.url
				});
		}
	});
var elm$http$Http$MySub = F2(
	function (a, b) {
		return {$: 'MySub', a: a, b: b};
	});
var elm$http$Http$subMap = F2(
	function (func, _n0) {
		var tracker = _n0.a;
		var toMsg = _n0.b;
		return A2(
			elm$http$Http$MySub,
			tracker,
			A2(elm$core$Basics$composeR, toMsg, func));
	});
_Platform_effectManagers['Http'] = _Platform_createManager(elm$http$Http$init, elm$http$Http$onEffects, elm$http$Http$onSelfMsg, elm$http$Http$cmdMap, elm$http$Http$subMap);
var elm$http$Http$command = _Platform_leaf('Http');
var elm$http$Http$subscription = _Platform_leaf('Http');
var elm$http$Http$request = function (r) {
	return elm$http$Http$command(
		elm$http$Http$Request(
			{allowCookiesFromOtherDomains: false, body: r.body, expect: r.expect, headers: r.headers, method: r.method, timeout: r.timeout, tracker: r.tracker, url: r.url}));
};
var elm$http$Http$post = function (r) {
	return elm$http$Http$request(
		{body: r.body, expect: r.expect, headers: _List_Nil, method: 'POST', timeout: elm$core$Maybe$Nothing, tracker: elm$core$Maybe$Nothing, url: r.url});
};
var author$project$Commands$doWhoami = function (msg) {
	return elm$http$Http$post(
		{
			body: elm$http$Http$emptyBody,
			expect: A2(elm$http$Http$expectJson, msg, author$project$Commands$decodeWhoami),
			url: '/api/v0/whoami'
		});
};
var author$project$Main$AdjustTimeZone = function (a) {
	return {$: 'AdjustTimeZone', a: a};
};
var author$project$Main$GotWhoamiResp = function (a) {
	return {$: 'GotWhoamiResp', a: a};
};
var author$project$Main$LoginLimbo = {$: 'LoginLimbo'};
var elm$core$Platform$Cmd$batch = _Platform_batch;
var elm$core$Task$Perform = function (a) {
	return {$: 'Perform', a: a};
};
var elm$core$Task$init = elm$core$Task$succeed(_Utils_Tuple0);
var elm$core$List$map = F2(
	function (f, xs) {
		return A3(
			elm$core$List$foldr,
			F2(
				function (x, acc) {
					return A2(
						elm$core$List$cons,
						f(x),
						acc);
				}),
			_List_Nil,
			xs);
	});
var elm$core$Task$map = F2(
	function (func, taskA) {
		return A2(
			elm$core$Task$andThen,
			function (a) {
				return elm$core$Task$succeed(
					func(a));
			},
			taskA);
	});
var elm$core$Task$spawnCmd = F2(
	function (router, _n0) {
		var task = _n0.a;
		return _Scheduler_spawn(
			A2(
				elm$core$Task$andThen,
				elm$core$Platform$sendToApp(router),
				task));
	});
var elm$core$Task$onEffects = F3(
	function (router, commands, state) {
		return A2(
			elm$core$Task$map,
			function (_n0) {
				return _Utils_Tuple0;
			},
			elm$core$Task$sequence(
				A2(
					elm$core$List$map,
					elm$core$Task$spawnCmd(router),
					commands)));
	});
var elm$core$Task$onSelfMsg = F3(
	function (_n0, _n1, _n2) {
		return elm$core$Task$succeed(_Utils_Tuple0);
	});
var elm$core$Task$cmdMap = F2(
	function (tagger, _n0) {
		var task = _n0.a;
		return elm$core$Task$Perform(
			A2(elm$core$Task$map, tagger, task));
	});
_Platform_effectManagers['Task'] = _Platform_createManager(elm$core$Task$init, elm$core$Task$onEffects, elm$core$Task$onSelfMsg, elm$core$Task$cmdMap);
var elm$core$Task$command = _Platform_leaf('Task');
var elm$core$Task$perform = F2(
	function (toMessage, task) {
		return elm$core$Task$command(
			elm$core$Task$Perform(
				A2(elm$core$Task$map, toMessage, task)));
	});
var elm$time$Time$Name = function (a) {
	return {$: 'Name', a: a};
};
var elm$time$Time$Offset = function (a) {
	return {$: 'Offset', a: a};
};
var elm$time$Time$Zone = F2(
	function (a, b) {
		return {$: 'Zone', a: a, b: b};
	});
var elm$time$Time$customZone = elm$time$Time$Zone;
var elm$time$Time$here = _Time_here(_Utils_Tuple0);
var elm$time$Time$utc = A2(elm$time$Time$Zone, 0, _List_Nil);
var author$project$Main$init = F3(
	function (_n0, url, key) {
		return _Utils_Tuple2(
			{key: key, loginState: author$project$Main$LoginLimbo, serverIsOnline: true, url: url, zone: elm$time$Time$utc},
			elm$core$Platform$Cmd$batch(
				_List_fromArray(
					[
						A2(elm$core$Task$perform, author$project$Main$AdjustTimeZone, elm$time$Time$here),
						author$project$Commands$doWhoami(author$project$Main$GotWhoamiResp)
					])));
	});
var author$project$Main$CommitsMsg = function (a) {
	return {$: 'CommitsMsg', a: a};
};
var author$project$Main$DeletedFilesMsg = function (a) {
	return {$: 'DeletedFilesMsg', a: a};
};
var author$project$Main$ListMsg = function (a) {
	return {$: 'ListMsg', a: a};
};
var author$project$Main$PingerIn = function (a) {
	return {$: 'PingerIn', a: a};
};
var author$project$Main$RemotesMsg = function (a) {
	return {$: 'RemotesMsg', a: a};
};
var author$project$Main$WebsocketIn = function (a) {
	return {$: 'WebsocketIn', a: a};
};
var author$project$Pinger$pinger = _Platform_incomingPort('pinger', elm$json$Json$Decode$string);
var author$project$Routes$Commits$OnScroll = function (a) {
	return {$: 'OnScroll', a: a};
};
var elm$json$Json$Decode$andThen = _Json_andThen;
var elm$json$Json$Decode$int = _Json_decodeInt;
var elm$json$Json$Decode$succeed = _Json_succeed;
var author$project$Scroll$scrollOrResize = _Platform_incomingPort(
	'scrollOrResize',
	A2(
		elm$json$Json$Decode$andThen,
		function (viewportWidth) {
			return A2(
				elm$json$Json$Decode$andThen,
				function (viewportHeight) {
					return A2(
						elm$json$Json$Decode$andThen,
						function (scrollTop) {
							return A2(
								elm$json$Json$Decode$andThen,
								function (pageHeight) {
									return elm$json$Json$Decode$succeed(
										{pageHeight: pageHeight, scrollTop: scrollTop, viewportHeight: viewportHeight, viewportWidth: viewportWidth});
								},
								A2(elm$json$Json$Decode$field, 'pageHeight', elm$json$Json$Decode$int));
						},
						A2(elm$json$Json$Decode$field, 'scrollTop', elm$json$Json$Decode$int));
				},
				A2(elm$json$Json$Decode$field, 'viewportHeight', elm$json$Json$Decode$int));
		},
		A2(elm$json$Json$Decode$field, 'viewportWidth', elm$json$Json$Decode$int)));
var author$project$Routes$Commits$subscriptions = function (model) {
	return author$project$Scroll$scrollOrResize(author$project$Routes$Commits$OnScroll);
};
var author$project$Routes$DeletedFiles$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Routes$DeletedFiles$OnScroll = function (a) {
	return {$: 'OnScroll', a: a};
};
var elm$core$Platform$Sub$batch = _Platform_batch;
var elm$browser$Browser$AnimationManager$Time = function (a) {
	return {$: 'Time', a: a};
};
var elm$browser$Browser$AnimationManager$State = F3(
	function (subs, request, oldTime) {
		return {oldTime: oldTime, request: request, subs: subs};
	});
var elm$browser$Browser$AnimationManager$init = elm$core$Task$succeed(
	A3(elm$browser$Browser$AnimationManager$State, _List_Nil, elm$core$Maybe$Nothing, 0));
var elm$browser$Browser$External = function (a) {
	return {$: 'External', a: a};
};
var elm$browser$Browser$Internal = function (a) {
	return {$: 'Internal', a: a};
};
var elm$browser$Browser$Dom$NotFound = function (a) {
	return {$: 'NotFound', a: a};
};
var elm$core$Basics$never = function (_n0) {
	never:
	while (true) {
		var nvr = _n0.a;
		var $temp$_n0 = nvr;
		_n0 = $temp$_n0;
		continue never;
	}
};
var elm$json$Json$Decode$map = _Json_map1;
var elm$json$Json$Decode$map2 = _Json_map2;
var elm$virtual_dom$VirtualDom$toHandlerInt = function (handler) {
	switch (handler.$) {
		case 'Normal':
			return 0;
		case 'MayStopPropagation':
			return 1;
		case 'MayPreventDefault':
			return 2;
		default:
			return 3;
	}
};
var elm$core$String$length = _String_length;
var elm$core$String$slice = _String_slice;
var elm$core$String$dropLeft = F2(
	function (n, string) {
		return (n < 1) ? string : A3(
			elm$core$String$slice,
			n,
			elm$core$String$length(string),
			string);
	});
var elm$core$String$startsWith = _String_startsWith;
var elm$url$Url$Http = {$: 'Http'};
var elm$url$Url$Https = {$: 'Https'};
var elm$core$String$indexes = _String_indexes;
var elm$core$String$isEmpty = function (string) {
	return string === '';
};
var elm$core$String$left = F2(
	function (n, string) {
		return (n < 1) ? '' : A3(elm$core$String$slice, 0, n, string);
	});
var elm$core$String$contains = _String_contains;
var elm$core$String$toInt = _String_toInt;
var elm$url$Url$Url = F6(
	function (protocol, host, port_, path, query, fragment) {
		return {fragment: fragment, host: host, path: path, port_: port_, protocol: protocol, query: query};
	});
var elm$url$Url$chompBeforePath = F5(
	function (protocol, path, params, frag, str) {
		if (elm$core$String$isEmpty(str) || A2(elm$core$String$contains, '@', str)) {
			return elm$core$Maybe$Nothing;
		} else {
			var _n0 = A2(elm$core$String$indexes, ':', str);
			if (!_n0.b) {
				return elm$core$Maybe$Just(
					A6(elm$url$Url$Url, protocol, str, elm$core$Maybe$Nothing, path, params, frag));
			} else {
				if (!_n0.b.b) {
					var i = _n0.a;
					var _n1 = elm$core$String$toInt(
						A2(elm$core$String$dropLeft, i + 1, str));
					if (_n1.$ === 'Nothing') {
						return elm$core$Maybe$Nothing;
					} else {
						var port_ = _n1;
						return elm$core$Maybe$Just(
							A6(
								elm$url$Url$Url,
								protocol,
								A2(elm$core$String$left, i, str),
								port_,
								path,
								params,
								frag));
					}
				} else {
					return elm$core$Maybe$Nothing;
				}
			}
		}
	});
var elm$url$Url$chompBeforeQuery = F4(
	function (protocol, params, frag, str) {
		if (elm$core$String$isEmpty(str)) {
			return elm$core$Maybe$Nothing;
		} else {
			var _n0 = A2(elm$core$String$indexes, '/', str);
			if (!_n0.b) {
				return A5(elm$url$Url$chompBeforePath, protocol, '/', params, frag, str);
			} else {
				var i = _n0.a;
				return A5(
					elm$url$Url$chompBeforePath,
					protocol,
					A2(elm$core$String$dropLeft, i, str),
					params,
					frag,
					A2(elm$core$String$left, i, str));
			}
		}
	});
var elm$url$Url$chompBeforeFragment = F3(
	function (protocol, frag, str) {
		if (elm$core$String$isEmpty(str)) {
			return elm$core$Maybe$Nothing;
		} else {
			var _n0 = A2(elm$core$String$indexes, '?', str);
			if (!_n0.b) {
				return A4(elm$url$Url$chompBeforeQuery, protocol, elm$core$Maybe$Nothing, frag, str);
			} else {
				var i = _n0.a;
				return A4(
					elm$url$Url$chompBeforeQuery,
					protocol,
					elm$core$Maybe$Just(
						A2(elm$core$String$dropLeft, i + 1, str)),
					frag,
					A2(elm$core$String$left, i, str));
			}
		}
	});
var elm$url$Url$chompAfterProtocol = F2(
	function (protocol, str) {
		if (elm$core$String$isEmpty(str)) {
			return elm$core$Maybe$Nothing;
		} else {
			var _n0 = A2(elm$core$String$indexes, '#', str);
			if (!_n0.b) {
				return A3(elm$url$Url$chompBeforeFragment, protocol, elm$core$Maybe$Nothing, str);
			} else {
				var i = _n0.a;
				return A3(
					elm$url$Url$chompBeforeFragment,
					protocol,
					elm$core$Maybe$Just(
						A2(elm$core$String$dropLeft, i + 1, str)),
					A2(elm$core$String$left, i, str));
			}
		}
	});
var elm$url$Url$fromString = function (str) {
	return A2(elm$core$String$startsWith, 'http://', str) ? A2(
		elm$url$Url$chompAfterProtocol,
		elm$url$Url$Http,
		A2(elm$core$String$dropLeft, 7, str)) : (A2(elm$core$String$startsWith, 'https://', str) ? A2(
		elm$url$Url$chompAfterProtocol,
		elm$url$Url$Https,
		A2(elm$core$String$dropLeft, 8, str)) : elm$core$Maybe$Nothing);
};
var elm$browser$Browser$AnimationManager$now = _Browser_now(_Utils_Tuple0);
var elm$browser$Browser$AnimationManager$rAF = _Browser_rAF(_Utils_Tuple0);
var elm$browser$Browser$AnimationManager$onEffects = F3(
	function (router, subs, _n0) {
		var request = _n0.request;
		var oldTime = _n0.oldTime;
		var _n1 = _Utils_Tuple2(request, subs);
		if (_n1.a.$ === 'Nothing') {
			if (!_n1.b.b) {
				var _n2 = _n1.a;
				return elm$browser$Browser$AnimationManager$init;
			} else {
				var _n4 = _n1.a;
				return A2(
					elm$core$Task$andThen,
					function (pid) {
						return A2(
							elm$core$Task$andThen,
							function (time) {
								return elm$core$Task$succeed(
									A3(
										elm$browser$Browser$AnimationManager$State,
										subs,
										elm$core$Maybe$Just(pid),
										time));
							},
							elm$browser$Browser$AnimationManager$now);
					},
					elm$core$Process$spawn(
						A2(
							elm$core$Task$andThen,
							elm$core$Platform$sendToSelf(router),
							elm$browser$Browser$AnimationManager$rAF)));
			}
		} else {
			if (!_n1.b.b) {
				var pid = _n1.a.a;
				return A2(
					elm$core$Task$andThen,
					function (_n3) {
						return elm$browser$Browser$AnimationManager$init;
					},
					elm$core$Process$kill(pid));
			} else {
				return elm$core$Task$succeed(
					A3(elm$browser$Browser$AnimationManager$State, subs, request, oldTime));
			}
		}
	});
var elm$time$Time$Posix = function (a) {
	return {$: 'Posix', a: a};
};
var elm$time$Time$millisToPosix = elm$time$Time$Posix;
var elm$browser$Browser$AnimationManager$onSelfMsg = F3(
	function (router, newTime, _n0) {
		var subs = _n0.subs;
		var oldTime = _n0.oldTime;
		var send = function (sub) {
			if (sub.$ === 'Time') {
				var tagger = sub.a;
				return A2(
					elm$core$Platform$sendToApp,
					router,
					tagger(
						elm$time$Time$millisToPosix(newTime)));
			} else {
				var tagger = sub.a;
				return A2(
					elm$core$Platform$sendToApp,
					router,
					tagger(newTime - oldTime));
			}
		};
		return A2(
			elm$core$Task$andThen,
			function (pid) {
				return A2(
					elm$core$Task$andThen,
					function (_n1) {
						return elm$core$Task$succeed(
							A3(
								elm$browser$Browser$AnimationManager$State,
								subs,
								elm$core$Maybe$Just(pid),
								newTime));
					},
					elm$core$Task$sequence(
						A2(elm$core$List$map, send, subs)));
			},
			elm$core$Process$spawn(
				A2(
					elm$core$Task$andThen,
					elm$core$Platform$sendToSelf(router),
					elm$browser$Browser$AnimationManager$rAF)));
	});
var elm$browser$Browser$AnimationManager$Delta = function (a) {
	return {$: 'Delta', a: a};
};
var elm$core$Basics$composeL = F3(
	function (g, f, x) {
		return g(
			f(x));
	});
var elm$browser$Browser$AnimationManager$subMap = F2(
	function (func, sub) {
		if (sub.$ === 'Time') {
			var tagger = sub.a;
			return elm$browser$Browser$AnimationManager$Time(
				A2(elm$core$Basics$composeL, func, tagger));
		} else {
			var tagger = sub.a;
			return elm$browser$Browser$AnimationManager$Delta(
				A2(elm$core$Basics$composeL, func, tagger));
		}
	});
_Platform_effectManagers['Browser.AnimationManager'] = _Platform_createManager(elm$browser$Browser$AnimationManager$init, elm$browser$Browser$AnimationManager$onEffects, elm$browser$Browser$AnimationManager$onSelfMsg, 0, elm$browser$Browser$AnimationManager$subMap);
var elm$browser$Browser$AnimationManager$subscription = _Platform_leaf('Browser.AnimationManager');
var elm$browser$Browser$AnimationManager$onAnimationFrame = function (tagger) {
	return elm$browser$Browser$AnimationManager$subscription(
		elm$browser$Browser$AnimationManager$Time(tagger));
};
var elm$browser$Browser$Events$onAnimationFrame = elm$browser$Browser$AnimationManager$onAnimationFrame;
var elm$core$Platform$Sub$none = elm$core$Platform$Sub$batch(_List_Nil);
var rundis$elm_bootstrap$Bootstrap$Alert$FadeClose = {$: 'FadeClose'};
var rundis$elm_bootstrap$Bootstrap$Alert$subscriptions = F2(
	function (visibility, animateMsg) {
		if (visibility.$ === 'StartClose') {
			return elm$browser$Browser$Events$onAnimationFrame(
				function (_n1) {
					return animateMsg(rundis$elm_bootstrap$Bootstrap$Alert$FadeClose);
				});
		} else {
			return elm$core$Platform$Sub$none;
		}
	});
var author$project$Routes$DeletedFiles$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				author$project$Scroll$scrollOrResize(author$project$Routes$DeletedFiles$OnScroll),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert.vis, author$project$Routes$DeletedFiles$AlertMsg)
			]));
};
var author$project$Modals$History$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$History$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$History$KeyPress = function (a) {
	return {$: 'KeyPress', a: a};
};
var elm$browser$Browser$Events$Document = {$: 'Document'};
var elm$browser$Browser$Events$MySub = F3(
	function (a, b, c) {
		return {$: 'MySub', a: a, b: b, c: c};
	});
var elm$browser$Browser$Events$State = F2(
	function (subs, pids) {
		return {pids: pids, subs: subs};
	});
var elm$browser$Browser$Events$init = elm$core$Task$succeed(
	A2(elm$browser$Browser$Events$State, _List_Nil, elm$core$Dict$empty));
var elm$browser$Browser$Events$nodeToKey = function (node) {
	if (node.$ === 'Document') {
		return 'd_';
	} else {
		return 'w_';
	}
};
var elm$browser$Browser$Events$addKey = function (sub) {
	var node = sub.a;
	var name = sub.b;
	return _Utils_Tuple2(
		_Utils_ap(
			elm$browser$Browser$Events$nodeToKey(node),
			name),
		sub);
};
var elm$browser$Browser$Events$Event = F2(
	function (key, event) {
		return {event: event, key: key};
	});
var elm$browser$Browser$Events$spawn = F3(
	function (router, key, _n0) {
		var node = _n0.a;
		var name = _n0.b;
		var actualNode = function () {
			if (node.$ === 'Document') {
				return _Browser_doc;
			} else {
				return _Browser_window;
			}
		}();
		return A2(
			elm$core$Task$map,
			function (value) {
				return _Utils_Tuple2(key, value);
			},
			A3(
				_Browser_on,
				actualNode,
				name,
				function (event) {
					return A2(
						elm$core$Platform$sendToSelf,
						router,
						A2(elm$browser$Browser$Events$Event, key, event));
				}));
	});
var elm$core$Dict$fromList = function (assocs) {
	return A3(
		elm$core$List$foldl,
		F2(
			function (_n0, dict) {
				var key = _n0.a;
				var value = _n0.b;
				return A3(elm$core$Dict$insert, key, value, dict);
			}),
		elm$core$Dict$empty,
		assocs);
};
var elm$core$Dict$foldl = F3(
	function (func, acc, dict) {
		foldl:
		while (true) {
			if (dict.$ === 'RBEmpty_elm_builtin') {
				return acc;
			} else {
				var key = dict.b;
				var value = dict.c;
				var left = dict.d;
				var right = dict.e;
				var $temp$func = func,
					$temp$acc = A3(
					func,
					key,
					value,
					A3(elm$core$Dict$foldl, func, acc, left)),
					$temp$dict = right;
				func = $temp$func;
				acc = $temp$acc;
				dict = $temp$dict;
				continue foldl;
			}
		}
	});
var elm$core$Dict$merge = F6(
	function (leftStep, bothStep, rightStep, leftDict, rightDict, initialResult) {
		var stepState = F3(
			function (rKey, rValue, _n0) {
				stepState:
				while (true) {
					var list = _n0.a;
					var result = _n0.b;
					if (!list.b) {
						return _Utils_Tuple2(
							list,
							A3(rightStep, rKey, rValue, result));
					} else {
						var _n2 = list.a;
						var lKey = _n2.a;
						var lValue = _n2.b;
						var rest = list.b;
						if (_Utils_cmp(lKey, rKey) < 0) {
							var $temp$rKey = rKey,
								$temp$rValue = rValue,
								$temp$_n0 = _Utils_Tuple2(
								rest,
								A3(leftStep, lKey, lValue, result));
							rKey = $temp$rKey;
							rValue = $temp$rValue;
							_n0 = $temp$_n0;
							continue stepState;
						} else {
							if (_Utils_cmp(lKey, rKey) > 0) {
								return _Utils_Tuple2(
									list,
									A3(rightStep, rKey, rValue, result));
							} else {
								return _Utils_Tuple2(
									rest,
									A4(bothStep, lKey, lValue, rValue, result));
							}
						}
					}
				}
			});
		var _n3 = A3(
			elm$core$Dict$foldl,
			stepState,
			_Utils_Tuple2(
				elm$core$Dict$toList(leftDict),
				initialResult),
			rightDict);
		var leftovers = _n3.a;
		var intermediateResult = _n3.b;
		return A3(
			elm$core$List$foldl,
			F2(
				function (_n4, result) {
					var k = _n4.a;
					var v = _n4.b;
					return A3(leftStep, k, v, result);
				}),
			intermediateResult,
			leftovers);
	});
var elm$core$Dict$union = F2(
	function (t1, t2) {
		return A3(elm$core$Dict$foldl, elm$core$Dict$insert, t2, t1);
	});
var elm$browser$Browser$Events$onEffects = F3(
	function (router, subs, state) {
		var stepRight = F3(
			function (key, sub, _n6) {
				var deads = _n6.a;
				var lives = _n6.b;
				var news = _n6.c;
				return _Utils_Tuple3(
					deads,
					lives,
					A2(
						elm$core$List$cons,
						A3(elm$browser$Browser$Events$spawn, router, key, sub),
						news));
			});
		var stepLeft = F3(
			function (_n4, pid, _n5) {
				var deads = _n5.a;
				var lives = _n5.b;
				var news = _n5.c;
				return _Utils_Tuple3(
					A2(elm$core$List$cons, pid, deads),
					lives,
					news);
			});
		var stepBoth = F4(
			function (key, pid, _n2, _n3) {
				var deads = _n3.a;
				var lives = _n3.b;
				var news = _n3.c;
				return _Utils_Tuple3(
					deads,
					A3(elm$core$Dict$insert, key, pid, lives),
					news);
			});
		var newSubs = A2(elm$core$List$map, elm$browser$Browser$Events$addKey, subs);
		var _n0 = A6(
			elm$core$Dict$merge,
			stepLeft,
			stepBoth,
			stepRight,
			state.pids,
			elm$core$Dict$fromList(newSubs),
			_Utils_Tuple3(_List_Nil, elm$core$Dict$empty, _List_Nil));
		var deadPids = _n0.a;
		var livePids = _n0.b;
		var makeNewPids = _n0.c;
		return A2(
			elm$core$Task$andThen,
			function (pids) {
				return elm$core$Task$succeed(
					A2(
						elm$browser$Browser$Events$State,
						newSubs,
						A2(
							elm$core$Dict$union,
							livePids,
							elm$core$Dict$fromList(pids))));
			},
			A2(
				elm$core$Task$andThen,
				function (_n1) {
					return elm$core$Task$sequence(makeNewPids);
				},
				elm$core$Task$sequence(
					A2(elm$core$List$map, elm$core$Process$kill, deadPids))));
	});
var elm$browser$Browser$Events$onSelfMsg = F3(
	function (router, _n0, state) {
		var key = _n0.key;
		var event = _n0.event;
		var toMessage = function (_n2) {
			var subKey = _n2.a;
			var _n3 = _n2.b;
			var node = _n3.a;
			var name = _n3.b;
			var decoder = _n3.c;
			return _Utils_eq(subKey, key) ? A2(_Browser_decodeEvent, decoder, event) : elm$core$Maybe$Nothing;
		};
		var messages = A2(elm$core$List$filterMap, toMessage, state.subs);
		return A2(
			elm$core$Task$andThen,
			function (_n1) {
				return elm$core$Task$succeed(state);
			},
			elm$core$Task$sequence(
				A2(
					elm$core$List$map,
					elm$core$Platform$sendToApp(router),
					messages)));
	});
var elm$browser$Browser$Events$subMap = F2(
	function (func, _n0) {
		var node = _n0.a;
		var name = _n0.b;
		var decoder = _n0.c;
		return A3(
			elm$browser$Browser$Events$MySub,
			node,
			name,
			A2(elm$json$Json$Decode$map, func, decoder));
	});
_Platform_effectManagers['Browser.Events'] = _Platform_createManager(elm$browser$Browser$Events$init, elm$browser$Browser$Events$onEffects, elm$browser$Browser$Events$onSelfMsg, 0, elm$browser$Browser$Events$subMap);
var elm$browser$Browser$Events$subscription = _Platform_leaf('Browser.Events');
var elm$browser$Browser$Events$on = F3(
	function (node, name, decoder) {
		return elm$browser$Browser$Events$subscription(
			A3(elm$browser$Browser$Events$MySub, node, name, decoder));
	});
var elm$browser$Browser$Events$onKeyPress = A2(elm$browser$Browser$Events$on, elm$browser$Browser$Events$Document, 'keypress');
var rundis$elm_bootstrap$Bootstrap$Modal$FadeClose = {$: 'FadeClose'};
var rundis$elm_bootstrap$Bootstrap$Modal$subscriptions = F2(
	function (visibility, animateMsg) {
		if (visibility.$ === 'StartClose') {
			return elm$browser$Browser$Events$onAnimationFrame(
				function (_n1) {
					return animateMsg(rundis$elm_bootstrap$Bootstrap$Modal$FadeClose);
				});
		} else {
			return elm$core$Platform$Sub$none;
		}
	});
var author$project$Modals$History$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$History$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$History$AlertMsg),
				elm$browser$Browser$Events$onKeyPress(
				A2(
					elm$json$Json$Decode$map,
					author$project$Modals$History$KeyPress,
					A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string)))
			]));
};
var author$project$Modals$Mkdir$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$Mkdir$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$Mkdir$KeyPress = F2(
	function (a, b) {
		return {$: 'KeyPress', a: a, b: b};
	});
var elm$core$List$filter = F2(
	function (isGood, list) {
		return A3(
			elm$core$List$foldr,
			F2(
				function (x, xs) {
					return isGood(x) ? A2(elm$core$List$cons, x, xs) : xs;
				}),
			_List_Nil,
			list);
	});
var author$project$Util$splitPath = function (path) {
	return A2(
		elm$core$List$filter,
		function (s) {
			return elm$core$String$length(s) > 0;
		},
		A2(elm$core$String$split, '/', path));
};
var author$project$Util$joinPath = function (paths) {
	return '/' + A2(
		elm$core$String$join,
		'/',
		A3(
			elm$core$List$foldr,
			elm$core$Basics$append,
			_List_Nil,
			A2(elm$core$List$map, author$project$Util$splitPath, paths)));
};
var elm$url$Url$percentDecode = _Url_percentDecode;
var author$project$Util$urlToPath = function (url) {
	var decodeUrlPart = function (e) {
		var _n1 = elm$url$Url$percentDecode(e);
		if (_n1.$ === 'Just') {
			var val = _n1.a;
			return val;
		} else {
			return '';
		}
	};
	var _n0 = author$project$Util$splitPath(url.path);
	if (!_n0.b) {
		return '/';
	} else {
		var xs = _n0.b;
		return '/' + A2(
			elm$core$String$join,
			'/',
			A2(elm$core$List$map, decodeUrlPart, xs));
	}
};
var author$project$Modals$Mkdir$pathFromUrl = F2(
	function (url, model) {
		return author$project$Util$joinPath(
			_List_fromArray(
				[
					author$project$Util$urlToPath(url),
					model.inputName
				]));
	});
var author$project$Modals$Mkdir$subscriptions = F2(
	function (url, model) {
		return elm$core$Platform$Sub$batch(
			_List_fromArray(
				[
					A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$Mkdir$AnimateModal),
					A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$Mkdir$AlertMsg),
					elm$browser$Browser$Events$onKeyPress(
					A2(
						elm$json$Json$Decode$map,
						author$project$Modals$Mkdir$KeyPress(
							A2(author$project$Modals$Mkdir$pathFromUrl, url, model)),
						A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string)))
				]));
	});
var author$project$Modals$MoveCopy$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$MoveCopy$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$MoveCopy$KeyPress = function (a) {
	return {$: 'KeyPress', a: a};
};
var author$project$Modals$MoveCopy$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$MoveCopy$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$MoveCopy$AlertMsg),
				elm$browser$Browser$Events$onKeyPress(
				A2(
					elm$json$Json$Decode$map,
					author$project$Modals$MoveCopy$KeyPress,
					A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string)))
			]));
};
var author$project$Modals$Remove$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$Remove$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$Remove$KeyPress = function (a) {
	return {$: 'KeyPress', a: a};
};
var author$project$Modals$Remove$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$Remove$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$Remove$AlertMsg),
				elm$browser$Browser$Events$onKeyPress(
				A2(
					elm$json$Json$Decode$map,
					author$project$Modals$Remove$KeyPress,
					A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string)))
			]));
};
var author$project$Modals$Rename$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$Rename$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$Rename$KeyPress = function (a) {
	return {$: 'KeyPress', a: a};
};
var author$project$Modals$Rename$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$Rename$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$Rename$AlertMsg),
				elm$browser$Browser$Events$onKeyPress(
				A2(
					elm$json$Json$Decode$map,
					author$project$Modals$Rename$KeyPress,
					A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string)))
			]));
};
var author$project$Modals$Share$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$Share$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$Share$AnimateModal)
			]));
};
var author$project$Modals$Upload$AlertMsg = F2(
	function (a, b) {
		return {$: 'AlertMsg', a: a, b: b};
	});
var author$project$Modals$Upload$UploadProgress = F2(
	function (a, b) {
		return {$: 'UploadProgress', a: a, b: b};
	});
var elm$http$Http$track = F2(
	function (tracker, toMsg) {
		return elm$http$Http$subscription(
			A2(elm$http$Http$MySub, tracker, toMsg));
	});
var author$project$Modals$Upload$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				elm$core$Platform$Sub$batch(
				A2(
					elm$core$List$map,
					function (p) {
						return A2(
							elm$http$Http$track,
							'upload-' + p,
							author$project$Modals$Upload$UploadProgress(p));
					},
					elm$core$Dict$keys(model.uploads))),
				elm$core$Platform$Sub$batch(
				A2(
					elm$core$List$map,
					function (a) {
						return A2(
							rundis$elm_bootstrap$Bootstrap$Alert$subscriptions,
							a.alert,
							author$project$Modals$Upload$AlertMsg(a.path));
					},
					model.success)),
				elm$core$Platform$Sub$batch(
				A2(
					elm$core$List$map,
					function (a) {
						return A2(
							rundis$elm_bootstrap$Bootstrap$Alert$subscriptions,
							a.alert,
							author$project$Modals$Upload$AlertMsg(a.path));
					},
					model.failed))
			]));
};
var author$project$Routes$Ls$ActionDropdownMsg = F2(
	function (a, b) {
		return {$: 'ActionDropdownMsg', a: a, b: b};
	});
var author$project$Routes$Ls$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Routes$Ls$CopyMsg = function (a) {
	return {$: 'CopyMsg', a: a};
};
var author$project$Routes$Ls$HistoryMsg = function (a) {
	return {$: 'HistoryMsg', a: a};
};
var author$project$Routes$Ls$MkdirMsg = function (a) {
	return {$: 'MkdirMsg', a: a};
};
var author$project$Routes$Ls$MoveMsg = function (a) {
	return {$: 'MoveMsg', a: a};
};
var author$project$Routes$Ls$RemoveMsg = function (a) {
	return {$: 'RemoveMsg', a: a};
};
var author$project$Routes$Ls$RenameMsg = function (a) {
	return {$: 'RenameMsg', a: a};
};
var author$project$Routes$Ls$ShareMsg = function (a) {
	return {$: 'ShareMsg', a: a};
};
var author$project$Routes$Ls$UploadMsg = function (a) {
	return {$: 'UploadMsg', a: a};
};
var elm$core$Platform$Sub$map = _Platform_map;
var elm$browser$Browser$Events$onClick = A2(elm$browser$Browser$Events$on, elm$browser$Browser$Events$Document, 'click');
var rundis$elm_bootstrap$Bootstrap$Dropdown$Closed = {$: 'Closed'};
var rundis$elm_bootstrap$Bootstrap$Dropdown$ListenClicks = {$: 'ListenClicks'};
var rundis$elm_bootstrap$Bootstrap$Dropdown$State = function (a) {
	return {$: 'State', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$updateStatus = F2(
	function (status, _n0) {
		var stateRec = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Dropdown$State(
			_Utils_update(
				stateRec,
				{status: status}));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$subscriptions = F2(
	function (state, toMsg) {
		var status = state.a.status;
		switch (status.$) {
			case 'Open':
				return elm$browser$Browser$Events$onAnimationFrame(
					function (_n1) {
						return toMsg(
							A2(rundis$elm_bootstrap$Bootstrap$Dropdown$updateStatus, rundis$elm_bootstrap$Bootstrap$Dropdown$ListenClicks, state));
					});
			case 'ListenClicks':
				return elm$browser$Browser$Events$onClick(
					elm$json$Json$Decode$succeed(
						toMsg(
							A2(rundis$elm_bootstrap$Bootstrap$Dropdown$updateStatus, rundis$elm_bootstrap$Bootstrap$Dropdown$Closed, state))));
			default:
				return elm$core$Platform$Sub$none;
		}
	});
var author$project$Routes$Ls$subscriptions = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		return elm$core$Platform$Sub$batch(
			_List_fromArray(
				[
					A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Routes$Ls$AlertMsg),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$HistoryMsg,
					author$project$Modals$History$subscriptions(model.historyState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$RenameMsg,
					author$project$Modals$Rename$subscriptions(model.renameState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$MoveMsg,
					author$project$Modals$MoveCopy$subscriptions(model.moveState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$CopyMsg,
					author$project$Modals$MoveCopy$subscriptions(model.copyState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$UploadMsg,
					author$project$Modals$Upload$subscriptions(model.uploadState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$MkdirMsg,
					A2(author$project$Modals$Mkdir$subscriptions, model.url, model.mkdirState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$RemoveMsg,
					author$project$Modals$Remove$subscriptions(model.removeState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Routes$Ls$ShareMsg,
					author$project$Modals$Share$subscriptions(model.shareState)),
					elm$core$Platform$Sub$batch(
					A2(
						elm$core$List$map,
						function (e) {
							return A2(
								rundis$elm_bootstrap$Bootstrap$Dropdown$subscriptions,
								e.dropdown,
								author$project$Routes$Ls$ActionDropdownMsg(e));
						},
						actualModel.entries))
				]));
	} else {
		return elm$core$Platform$Sub$none;
	}
};
var author$project$Modals$RemoteAdd$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$RemoteAdd$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$RemoteAdd$ConflictDropdownMsg = function (a) {
	return {$: 'ConflictDropdownMsg', a: a};
};
var author$project$Modals$RemoteAdd$KeyPress = function (a) {
	return {$: 'KeyPress', a: a};
};
var author$project$Modals$RemoteAdd$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$RemoteAdd$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$RemoteAdd$AlertMsg),
				elm$browser$Browser$Events$onKeyPress(
				A2(
					elm$json$Json$Decode$map,
					author$project$Modals$RemoteAdd$KeyPress,
					A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string))),
				A2(rundis$elm_bootstrap$Bootstrap$Dropdown$subscriptions, model.conflictDropdown, author$project$Modals$RemoteAdd$ConflictDropdownMsg)
			]));
};
var author$project$Modals$RemoteFolders$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$RemoteFolders$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$RemoteFolders$ConflictDropdownMsg = F2(
	function (a, b) {
		return {$: 'ConflictDropdownMsg', a: a, b: b};
	});
var author$project$Modals$RemoteFolders$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$RemoteFolders$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$RemoteFolders$AlertMsg),
				elm$core$Platform$Sub$batch(
				A2(
					elm$core$List$map,
					function (_n0) {
						var name = _n0.a;
						var state = _n0.b;
						return A2(
							rundis$elm_bootstrap$Bootstrap$Dropdown$subscriptions,
							state,
							author$project$Modals$RemoteFolders$ConflictDropdownMsg(name));
					},
					elm$core$Dict$toList(model.conflictDropdowns)))
			]));
};
var author$project$Modals$RemoteRemove$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Modals$RemoteRemove$AnimateModal = function (a) {
	return {$: 'AnimateModal', a: a};
};
var author$project$Modals$RemoteRemove$KeyPress = function (a) {
	return {$: 'KeyPress', a: a};
};
var author$project$Modals$RemoteRemove$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Modal$subscriptions, model.modal, author$project$Modals$RemoteRemove$AnimateModal),
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert, author$project$Modals$RemoteRemove$AlertMsg),
				elm$browser$Browser$Events$onKeyPress(
				A2(
					elm$json$Json$Decode$map,
					author$project$Modals$RemoteRemove$KeyPress,
					A2(elm$json$Json$Decode$field, 'key', elm$json$Json$Decode$string)))
			]));
};
var author$project$Routes$Remotes$ActionDropdownMsg = F2(
	function (a, b) {
		return {$: 'ActionDropdownMsg', a: a, b: b};
	});
var author$project$Routes$Remotes$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Routes$Remotes$ConflictDropdownMsg = F2(
	function (a, b) {
		return {$: 'ConflictDropdownMsg', a: a, b: b};
	});
var author$project$Routes$Remotes$RemoteAddMsg = function (a) {
	return {$: 'RemoteAddMsg', a: a};
};
var author$project$Routes$Remotes$RemoteFolderMsg = function (a) {
	return {$: 'RemoteFolderMsg', a: a};
};
var author$project$Routes$Remotes$RemoteRemoveMsg = function (a) {
	return {$: 'RemoteRemoveMsg', a: a};
};
var author$project$Routes$Remotes$subscriptions = function (model) {
	return elm$core$Platform$Sub$batch(
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Alert$subscriptions, model.alert.vis, author$project$Routes$Remotes$AlertMsg),
				A2(
				elm$core$Platform$Sub$map,
				author$project$Routes$Remotes$RemoteAddMsg,
				author$project$Modals$RemoteAdd$subscriptions(model.remoteAddState)),
				A2(
				elm$core$Platform$Sub$map,
				author$project$Routes$Remotes$RemoteRemoveMsg,
				author$project$Modals$RemoteRemove$subscriptions(model.remoteRemoveState)),
				A2(
				elm$core$Platform$Sub$map,
				author$project$Routes$Remotes$RemoteFolderMsg,
				author$project$Modals$RemoteFolders$subscriptions(model.remoteFoldersState)),
				elm$core$Platform$Sub$batch(
				A2(
					elm$core$List$map,
					function (_n0) {
						var name = _n0.a;
						var state = _n0.b;
						return A2(
							rundis$elm_bootstrap$Bootstrap$Dropdown$subscriptions,
							state,
							author$project$Routes$Remotes$ActionDropdownMsg(name));
					},
					elm$core$Dict$toList(model.actionDropdowns))),
				elm$core$Platform$Sub$batch(
				A2(
					elm$core$List$map,
					function (_n1) {
						var name = _n1.a;
						var state = _n1.b;
						return A2(
							rundis$elm_bootstrap$Bootstrap$Dropdown$subscriptions,
							state,
							author$project$Routes$Remotes$ConflictDropdownMsg(name));
					},
					elm$core$Dict$toList(model.conflictDropdowns)))
			]));
};
var author$project$Websocket$incoming = _Platform_incomingPort('incoming', elm$json$Json$Decode$string);
var author$project$Main$subscriptions = function (model) {
	var _n0 = model.loginState;
	if (_n0.$ === 'LoginSuccess') {
		var viewState = _n0.a;
		return elm$core$Platform$Sub$batch(
			_List_fromArray(
				[
					A2(
					elm$core$Platform$Sub$map,
					author$project$Main$ListMsg,
					author$project$Routes$Ls$subscriptions(viewState.listState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Main$CommitsMsg,
					author$project$Routes$Commits$subscriptions(viewState.commitsState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Main$RemotesMsg,
					author$project$Routes$Remotes$subscriptions(viewState.remoteState)),
					A2(
					elm$core$Platform$Sub$map,
					author$project$Main$DeletedFilesMsg,
					author$project$Routes$DeletedFiles$subscriptions(viewState.deletedFilesState)),
					author$project$Websocket$incoming(author$project$Main$WebsocketIn),
					author$project$Pinger$pinger(author$project$Main$PingerIn)
				]));
	} else {
		return elm$core$Platform$Sub$none;
	}
};
var author$project$Commands$LoginQuery = F2(
	function (username, password) {
		return {password: password, username: username};
	});
var author$project$Commands$LoginResponse = F4(
	function (username, rights, isAnon, anonIsAllowed) {
		return {anonIsAllowed: anonIsAllowed, isAnon: isAnon, rights: rights, username: username};
	});
var elm$json$Json$Decode$map4 = _Json_map4;
var author$project$Commands$decodeLoginResponse = A5(
	elm$json$Json$Decode$map4,
	author$project$Commands$LoginResponse,
	A2(elm$json$Json$Decode$field, 'username', elm$json$Json$Decode$string),
	A2(
		elm$json$Json$Decode$field,
		'rights',
		elm$json$Json$Decode$list(elm$json$Json$Decode$string)),
	A2(elm$json$Json$Decode$field, 'is_anon', elm$json$Json$Decode$bool),
	A2(elm$json$Json$Decode$field, 'anon_is_allowed', elm$json$Json$Decode$bool));
var elm$json$Json$Encode$object = function (pairs) {
	return _Json_wrap(
		A3(
			elm$core$List$foldl,
			F2(
				function (_n0, obj) {
					var k = _n0.a;
					var v = _n0.b;
					return A3(_Json_addField, k, v, obj);
				}),
			_Json_emptyObject(_Utils_Tuple0),
			pairs));
};
var elm$json$Json$Encode$string = _Json_wrap;
var author$project$Commands$encodeLoginQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'username',
				elm$json$Json$Encode$string(q.username)),
				_Utils_Tuple2(
				'password',
				elm$json$Json$Encode$string(q.password))
			]));
};
var elm$http$Http$jsonBody = function (value) {
	return A2(
		_Http_pair,
		'application/json',
		A2(elm$json$Json$Encode$encode, 0, value));
};
var author$project$Commands$doLogin = F3(
	function (toMsg, user, pass) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeLoginQuery(
						A2(author$project$Commands$LoginQuery, user, pass))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeLoginResponse),
				url: '/api/v0/login'
			});
	});
var author$project$Commands$doLogout = function (msg) {
	return elm$http$Http$post(
		{
			body: elm$http$Http$emptyBody,
			expect: A2(
				elm$http$Http$expectJson,
				msg,
				A2(elm$json$Json$Decode$field, 'success', elm$json$Json$Decode$bool)),
			url: '/api/v0/logout'
		});
};
var author$project$Main$DiffMsg = function (a) {
	return {$: 'DiffMsg', a: a};
};
var author$project$Main$GotLoginResp = function (a) {
	return {$: 'GotLoginResp', a: a};
};
var author$project$Main$GotLogoutResp = F2(
	function (a, b) {
		return {$: 'GotLogoutResp', a: a, b: b};
	});
var author$project$Main$LoginFailure = F3(
	function (a, b, c) {
		return {$: 'LoginFailure', a: a, b: b, c: c};
	});
var author$project$Main$LoginLoading = F2(
	function (a, b) {
		return {$: 'LoginLoading', a: a, b: b};
	});
var author$project$Main$LoginReady = F2(
	function (a, b) {
		return {$: 'LoginReady', a: a, b: b};
	});
var author$project$Main$LoginSuccess = function (a) {
	return {$: 'LoginSuccess', a: a};
};
var author$project$Main$ViewCommits = {$: 'ViewCommits'};
var author$project$Main$ViewDeletedFiles = {$: 'ViewDeletedFiles'};
var author$project$Main$ViewDiff = {$: 'ViewDiff'};
var author$project$Main$ViewList = {$: 'ViewList'};
var author$project$Main$ViewNotFound = {$: 'ViewNotFound'};
var author$project$Main$ViewRemotes = {$: 'ViewRemotes'};
var elm$core$List$drop = F2(
	function (n, list) {
		drop:
		while (true) {
			if (n <= 0) {
				return list;
			} else {
				if (!list.b) {
					return list;
				} else {
					var x = list.a;
					var xs = list.b;
					var $temp$n = n - 1,
						$temp$list = xs;
					n = $temp$n;
					list = $temp$list;
					continue drop;
				}
			}
		}
	});
var elm$core$List$head = function (list) {
	if (list.b) {
		var x = list.a;
		var xs = list.b;
		return elm$core$Maybe$Just(x);
	} else {
		return elm$core$Maybe$Nothing;
	}
};
var elm$core$List$any = F2(
	function (isOkay, list) {
		any:
		while (true) {
			if (!list.b) {
				return false;
			} else {
				var x = list.a;
				var xs = list.b;
				if (isOkay(x)) {
					return true;
				} else {
					var $temp$isOkay = isOkay,
						$temp$list = xs;
					isOkay = $temp$isOkay;
					list = $temp$list;
					continue any;
				}
			}
		}
	});
var elm$core$List$member = F2(
	function (x, xs) {
		return A2(
			elm$core$List$any,
			function (a) {
				return _Utils_eq(a, x);
			},
			xs);
	});
var author$project$Main$viewFromUrl = F2(
	function (rights, url) {
		var _n0 = elm$core$List$head(
			A2(
				elm$core$List$drop,
				1,
				A2(elm$core$String$split, '/', url.path)));
		if (_n0.$ === 'Nothing') {
			return author$project$Main$ViewNotFound;
		} else {
			var first = _n0.a;
			switch (first) {
				case 'view':
					return author$project$Main$ViewList;
				case 'log':
					return author$project$Main$ViewCommits;
				case 'remotes':
					return author$project$Main$ViewRemotes;
				case 'deleted':
					return author$project$Main$ViewDeletedFiles;
				case 'diff':
					return author$project$Main$ViewDiff;
				case '':
					return A2(elm$core$List$member, 'fs.view', rights) ? author$project$Main$ViewList : author$project$Main$ViewRemotes;
				default:
					return author$project$Main$ViewNotFound;
			}
		}
	});
var author$project$Routes$Commits$Loading = {$: 'Loading'};
var author$project$Util$Info = {$: 'Info'};
var rundis$elm_bootstrap$Bootstrap$Alert$Closed = {$: 'Closed'};
var rundis$elm_bootstrap$Bootstrap$Alert$closed = rundis$elm_bootstrap$Bootstrap$Alert$Closed;
var author$project$Util$defaultAlertState = {message: '', typ: author$project$Util$Info, vis: rundis$elm_bootstrap$Bootstrap$Alert$closed};
var author$project$Routes$Commits$newModel = F4(
	function (url, key, zone, rights) {
		return {alert: author$project$Util$defaultAlertState, filter: '', haveStagedChanges: false, key: key, offset: 0, rights: rights, state: author$project$Routes$Commits$Loading, url: url, zone: zone};
	});
var author$project$Commands$LogQuery = F3(
	function (offset, limit, filter) {
		return {filter: filter, limit: limit, offset: offset};
	});
var author$project$Commands$Log = F2(
	function (haveStagedChanges, commits) {
		return {commits: commits, haveStagedChanges: haveStagedChanges};
	});
var NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$custom = elm$json$Json$Decode$map2(elm$core$Basics$apR);
var NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required = F3(
	function (key, valDecoder, decoder) {
		return A2(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$custom,
			A2(elm$json$Json$Decode$field, key, valDecoder),
			decoder);
	});
var author$project$Commands$Commit = F5(
	function (date, msg, tags, hash, index) {
		return {date: date, hash: hash, index: index, msg: msg, tags: tags};
	});
var author$project$Commands$timestampToPosix = A2(
	elm$json$Json$Decode$andThen,
	function (ms) {
		return elm$json$Json$Decode$succeed(
			elm$time$Time$millisToPosix(ms));
	},
	elm$json$Json$Decode$int);
var author$project$Commands$decodeCommit = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'index',
	elm$json$Json$Decode$int,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'hash',
		elm$json$Json$Decode$string,
		A3(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
			'tags',
			elm$json$Json$Decode$list(elm$json$Json$Decode$string),
			A3(
				NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
				'msg',
				elm$json$Json$Decode$string,
				A3(
					NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
					'date',
					author$project$Commands$timestampToPosix,
					elm$json$Json$Decode$succeed(author$project$Commands$Commit))))));
var author$project$Commands$decodeLog = A3(
	elm$json$Json$Decode$map2,
	author$project$Commands$Log,
	A2(elm$json$Json$Decode$field, 'have_staged_changes', elm$json$Json$Decode$bool),
	A2(
		elm$json$Json$Decode$field,
		'commits',
		elm$json$Json$Decode$list(author$project$Commands$decodeCommit)));
var elm$json$Json$Encode$int = _Json_wrap;
var author$project$Commands$encodeLog = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'offset',
				elm$json$Json$Encode$int(q.offset)),
				_Utils_Tuple2(
				'limit',
				elm$json$Json$Encode$int(q.limit)),
				_Utils_Tuple2(
				'filter',
				elm$json$Json$Encode$string(q.filter))
			]));
};
var author$project$Commands$doLog = F4(
	function (msg, offset, limit, filter) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeLog(
						A3(author$project$Commands$LogQuery, offset, limit, filter))),
				expect: A2(elm$http$Http$expectJson, msg, author$project$Commands$decodeLog),
				url: '/api/v0/log'
			});
	});
var author$project$Routes$Commits$GotLogResponse = F2(
	function (a, b) {
		return {$: 'GotLogResponse', a: a, b: b};
	});
var author$project$Routes$Commits$loadLimit = 20;
var author$project$Routes$Commits$reload = function (model) {
	return A4(
		author$project$Commands$doLog,
		author$project$Routes$Commits$GotLogResponse(true),
		0,
		model.offset + author$project$Routes$Commits$loadLimit,
		model.filter);
};
var author$project$Routes$DeletedFiles$Loading = {$: 'Loading'};
var author$project$Routes$DeletedFiles$newModel = F4(
	function (url, key, zone, rights) {
		return {alert: author$project$Util$defaultAlertState, filter: '', key: key, offset: 0, rights: rights, state: author$project$Routes$DeletedFiles$Loading, url: url, zone: zone};
	});
var author$project$Commands$DeletedFilesQuery = F3(
	function (offset, limit, filter) {
		return {filter: filter, limit: limit, offset: offset};
	});
var author$project$Commands$Entry = function (dropdown) {
	return function (path) {
		return function (user) {
			return function (size) {
				return function (inode) {
					return function (depth) {
						return function (lastModified) {
							return function (isDir) {
								return function (isPinned) {
									return function (isExplicit) {
										return {depth: depth, dropdown: dropdown, inode: inode, isDir: isDir, isExplicit: isExplicit, isPinned: isPinned, lastModified: lastModified, path: path, size: size, user: user};
									};
								};
							};
						};
					};
				};
			};
		};
	};
};
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$Area = F4(
	function (top, left, width, height) {
		return {height: height, left: left, top: top, width: width};
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$initialState = rundis$elm_bootstrap$Bootstrap$Dropdown$State(
	{
		menuSize: A4(rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$Area, 0, 0, 0, 0),
		status: rundis$elm_bootstrap$Bootstrap$Dropdown$Closed,
		toggleSize: A4(rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$Area, 0, 0, 0, 0)
	});
var author$project$Commands$decodeEntry = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'is_explicit',
	elm$json$Json$Decode$bool,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'is_pinned',
		elm$json$Json$Decode$bool,
		A3(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
			'is_dir',
			elm$json$Json$Decode$bool,
			A3(
				NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
				'last_modified_ms',
				author$project$Commands$timestampToPosix,
				A3(
					NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
					'depth',
					elm$json$Json$Decode$int,
					A3(
						NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
						'inode',
						elm$json$Json$Decode$int,
						A3(
							NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
							'size',
							elm$json$Json$Decode$int,
							A3(
								NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
								'user',
								elm$json$Json$Decode$string,
								A3(
									NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
									'path',
									elm$json$Json$Decode$string,
									elm$json$Json$Decode$succeed(
										author$project$Commands$Entry(rundis$elm_bootstrap$Bootstrap$Dropdown$initialState)))))))))));
var author$project$Commands$decodeDeletedFiles = A2(
	elm$json$Json$Decode$field,
	'entries',
	elm$json$Json$Decode$list(author$project$Commands$decodeEntry));
var author$project$Commands$encodeDeletedFiles = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'offset',
				elm$json$Json$Encode$int(q.offset)),
				_Utils_Tuple2(
				'limit',
				elm$json$Json$Encode$int(q.limit)),
				_Utils_Tuple2(
				'filter',
				elm$json$Json$Encode$string(q.filter))
			]));
};
var author$project$Commands$doDeletedFiles = F4(
	function (msg, offset, limit, filter) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeDeletedFiles(
						A3(author$project$Commands$DeletedFilesQuery, offset, limit, filter))),
				expect: A2(elm$http$Http$expectJson, msg, author$project$Commands$decodeDeletedFiles),
				url: '/api/v0/deleted'
			});
	});
var author$project$Routes$DeletedFiles$GotDeletedPathsResponse = F2(
	function (a, b) {
		return {$: 'GotDeletedPathsResponse', a: a, b: b};
	});
var author$project$Routes$DeletedFiles$loadLimit = 25;
var author$project$Routes$DeletedFiles$reload = function (model) {
	return A4(
		author$project$Commands$doDeletedFiles,
		author$project$Routes$DeletedFiles$GotDeletedPathsResponse(true),
		0,
		model.offset + author$project$Routes$DeletedFiles$loadLimit,
		model.filter);
};
var author$project$Routes$Diff$Loading = {$: 'Loading'};
var author$project$Routes$Diff$newModel = F3(
	function (key, url, zone) {
		return {key: key, state: author$project$Routes$Diff$Loading, url: url, zone: zone};
	});
var author$project$Commands$RemoteDiffQuery = function (name) {
	return {name: name};
};
var author$project$Commands$Diff = F7(
	function (added, removed, ignored, missing, moved, merged, conflict) {
		return {added: added, conflict: conflict, ignored: ignored, merged: merged, missing: missing, moved: moved, removed: removed};
	});
var author$project$Commands$DiffPair = F2(
	function (src, dst) {
		return {dst: dst, src: src};
	});
var author$project$Commands$decodeDiffPair = A3(
	elm$json$Json$Decode$map2,
	author$project$Commands$DiffPair,
	A2(elm$json$Json$Decode$field, 'src', author$project$Commands$decodeEntry),
	A2(elm$json$Json$Decode$field, 'dst', author$project$Commands$decodeEntry));
var author$project$Commands$decodeDiff = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'conflict',
	elm$json$Json$Decode$list(author$project$Commands$decodeDiffPair),
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'merged',
		elm$json$Json$Decode$list(author$project$Commands$decodeDiffPair),
		A3(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
			'moved',
			elm$json$Json$Decode$list(author$project$Commands$decodeDiffPair),
			A3(
				NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
				'missing',
				elm$json$Json$Decode$list(author$project$Commands$decodeEntry),
				A3(
					NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
					'ignored',
					elm$json$Json$Decode$list(author$project$Commands$decodeEntry),
					A3(
						NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
						'removed',
						elm$json$Json$Decode$list(author$project$Commands$decodeEntry),
						A3(
							NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
							'added',
							elm$json$Json$Decode$list(author$project$Commands$decodeEntry),
							elm$json$Json$Decode$succeed(author$project$Commands$Diff))))))));
var author$project$Commands$decodeDiffResponse = A2(elm$json$Json$Decode$field, 'diff', author$project$Commands$decodeDiff);
var author$project$Commands$encodeRemoteDiffQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'name',
				elm$json$Json$Encode$string(q.name))
			]));
};
var author$project$Commands$doRemoteDiff = F2(
	function (toMsg, name) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeRemoteDiffQuery(
						author$project$Commands$RemoteDiffQuery(name))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeDiffResponse),
				url: '/api/v0/remotes/diff'
			});
	});
var author$project$Routes$Diff$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var elm$core$Maybe$withDefault = F2(
	function (_default, maybe) {
		if (maybe.$ === 'Just') {
			var value = maybe.a;
			return value;
		} else {
			return _default;
		}
	});
var elm$url$Url$Parser$State = F5(
	function (visited, unvisited, params, frag, value) {
		return {frag: frag, params: params, unvisited: unvisited, value: value, visited: visited};
	});
var elm$url$Url$Parser$getFirstMatch = function (states) {
	getFirstMatch:
	while (true) {
		if (!states.b) {
			return elm$core$Maybe$Nothing;
		} else {
			var state = states.a;
			var rest = states.b;
			var _n1 = state.unvisited;
			if (!_n1.b) {
				return elm$core$Maybe$Just(state.value);
			} else {
				if ((_n1.a === '') && (!_n1.b.b)) {
					return elm$core$Maybe$Just(state.value);
				} else {
					var $temp$states = rest;
					states = $temp$states;
					continue getFirstMatch;
				}
			}
		}
	}
};
var elm$url$Url$Parser$removeFinalEmpty = function (segments) {
	if (!segments.b) {
		return _List_Nil;
	} else {
		if ((segments.a === '') && (!segments.b.b)) {
			return _List_Nil;
		} else {
			var segment = segments.a;
			var rest = segments.b;
			return A2(
				elm$core$List$cons,
				segment,
				elm$url$Url$Parser$removeFinalEmpty(rest));
		}
	}
};
var elm$url$Url$Parser$preparePath = function (path) {
	var _n0 = A2(elm$core$String$split, '/', path);
	if (_n0.b && (_n0.a === '')) {
		var segments = _n0.b;
		return elm$url$Url$Parser$removeFinalEmpty(segments);
	} else {
		var segments = _n0;
		return elm$url$Url$Parser$removeFinalEmpty(segments);
	}
};
var elm$url$Url$Parser$addToParametersHelp = F2(
	function (value, maybeList) {
		if (maybeList.$ === 'Nothing') {
			return elm$core$Maybe$Just(
				_List_fromArray(
					[value]));
		} else {
			var list = maybeList.a;
			return elm$core$Maybe$Just(
				A2(elm$core$List$cons, value, list));
		}
	});
var elm$url$Url$Parser$addParam = F2(
	function (segment, dict) {
		var _n0 = A2(elm$core$String$split, '=', segment);
		if ((_n0.b && _n0.b.b) && (!_n0.b.b.b)) {
			var rawKey = _n0.a;
			var _n1 = _n0.b;
			var rawValue = _n1.a;
			var _n2 = elm$url$Url$percentDecode(rawKey);
			if (_n2.$ === 'Nothing') {
				return dict;
			} else {
				var key = _n2.a;
				var _n3 = elm$url$Url$percentDecode(rawValue);
				if (_n3.$ === 'Nothing') {
					return dict;
				} else {
					var value = _n3.a;
					return A3(
						elm$core$Dict$update,
						key,
						elm$url$Url$Parser$addToParametersHelp(value),
						dict);
				}
			}
		} else {
			return dict;
		}
	});
var elm$url$Url$Parser$prepareQuery = function (maybeQuery) {
	if (maybeQuery.$ === 'Nothing') {
		return elm$core$Dict$empty;
	} else {
		var qry = maybeQuery.a;
		return A3(
			elm$core$List$foldr,
			elm$url$Url$Parser$addParam,
			elm$core$Dict$empty,
			A2(elm$core$String$split, '&', qry));
	}
};
var elm$url$Url$Parser$parse = F2(
	function (_n0, url) {
		var parser = _n0.a;
		return elm$url$Url$Parser$getFirstMatch(
			parser(
				A5(
					elm$url$Url$Parser$State,
					_List_Nil,
					elm$url$Url$Parser$preparePath(url.path),
					elm$url$Url$Parser$prepareQuery(url.query),
					url.fragment,
					elm$core$Basics$identity)));
	});
var elm$url$Url$Parser$Parser = function (a) {
	return {$: 'Parser', a: a};
};
var elm$url$Url$Parser$s = function (str) {
	return elm$url$Url$Parser$Parser(
		function (_n0) {
			var visited = _n0.visited;
			var unvisited = _n0.unvisited;
			var params = _n0.params;
			var frag = _n0.frag;
			var value = _n0.value;
			if (!unvisited.b) {
				return _List_Nil;
			} else {
				var next = unvisited.a;
				var rest = unvisited.b;
				return _Utils_eq(next, str) ? _List_fromArray(
					[
						A5(
						elm$url$Url$Parser$State,
						A2(elm$core$List$cons, next, visited),
						rest,
						params,
						frag,
						value)
					]) : _List_Nil;
			}
		});
};
var elm$core$List$append = F2(
	function (xs, ys) {
		if (!ys.b) {
			return xs;
		} else {
			return A3(elm$core$List$foldr, elm$core$List$cons, ys, xs);
		}
	});
var elm$core$List$concat = function (lists) {
	return A3(elm$core$List$foldr, elm$core$List$append, _List_Nil, lists);
};
var elm$core$List$concatMap = F2(
	function (f, list) {
		return elm$core$List$concat(
			A2(elm$core$List$map, f, list));
	});
var elm$url$Url$Parser$slash = F2(
	function (_n0, _n1) {
		var parseBefore = _n0.a;
		var parseAfter = _n1.a;
		return elm$url$Url$Parser$Parser(
			function (state) {
				return A2(
					elm$core$List$concatMap,
					parseAfter,
					parseBefore(state));
			});
	});
var elm$url$Url$Parser$custom = F2(
	function (tipe, stringToSomething) {
		return elm$url$Url$Parser$Parser(
			function (_n0) {
				var visited = _n0.visited;
				var unvisited = _n0.unvisited;
				var params = _n0.params;
				var frag = _n0.frag;
				var value = _n0.value;
				if (!unvisited.b) {
					return _List_Nil;
				} else {
					var next = unvisited.a;
					var rest = unvisited.b;
					var _n2 = stringToSomething(next);
					if (_n2.$ === 'Just') {
						var nextValue = _n2.a;
						return _List_fromArray(
							[
								A5(
								elm$url$Url$Parser$State,
								A2(elm$core$List$cons, next, visited),
								rest,
								params,
								frag,
								value(nextValue))
							]);
					} else {
						return _List_Nil;
					}
				}
			});
	});
var elm$url$Url$Parser$string = A2(elm$url$Url$Parser$custom, 'STRING', elm$core$Maybe$Just);
var author$project$Routes$Diff$nameFromUrl = function (url) {
	return A2(
		elm$core$Maybe$withDefault,
		'',
		A2(
			elm$url$Url$Parser$parse,
			A2(
				elm$url$Url$Parser$slash,
				elm$url$Url$Parser$s('diff'),
				elm$url$Url$Parser$string),
			url));
};
var elm$core$Platform$Cmd$none = elm$core$Platform$Cmd$batch(_List_Nil);
var author$project$Routes$Diff$reload = F2(
	function (model, url) {
		var remoteName = author$project$Routes$Diff$nameFromUrl(url);
		return (elm$core$String$length(remoteName) > 0) ? A2(author$project$Commands$doRemoteDiff, author$project$Routes$Diff$GotResponse, remoteName) : elm$core$Platform$Cmd$none;
	});
var author$project$Commands$ListQuery = F2(
	function (root, filter) {
		return {filter: filter, root: root};
	});
var author$project$Commands$ListResponse = F3(
	function (self, isFiltered, entries) {
		return {entries: entries, isFiltered: isFiltered, self: self};
	});
var elm$json$Json$Decode$map3 = _Json_map3;
var author$project$Commands$decodeListResponse = A4(
	elm$json$Json$Decode$map3,
	author$project$Commands$ListResponse,
	A2(elm$json$Json$Decode$field, 'self', author$project$Commands$decodeEntry),
	A2(elm$json$Json$Decode$field, 'is_filtered', elm$json$Json$Decode$bool),
	A2(
		elm$json$Json$Decode$field,
		'files',
		elm$json$Json$Decode$list(author$project$Commands$decodeEntry)));
var author$project$Commands$encodeListResponse = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'root',
				elm$json$Json$Encode$string(q.root)),
				_Utils_Tuple2(
				'filter',
				elm$json$Json$Encode$string(q.filter))
			]));
};
var author$project$Commands$doListQuery = F3(
	function (toMsg, path, filter) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeListResponse(
						A2(author$project$Commands$ListQuery, path, filter))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeListResponse),
				url: '/api/v0/ls'
			});
	});
var author$project$Routes$Ls$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var elm$url$Url$Parser$query = function (_n0) {
	var queryParser = _n0.a;
	return elm$url$Url$Parser$Parser(
		function (_n1) {
			var visited = _n1.visited;
			var unvisited = _n1.unvisited;
			var params = _n1.params;
			var frag = _n1.frag;
			var value = _n1.value;
			return _List_fromArray(
				[
					A5(
					elm$url$Url$Parser$State,
					visited,
					unvisited,
					params,
					frag,
					value(
						queryParser(params)))
				]);
		});
};
var elm$url$Url$Parser$Internal$Parser = function (a) {
	return {$: 'Parser', a: a};
};
var elm$url$Url$Parser$Query$map = F2(
	function (func, _n0) {
		var a = _n0.a;
		return elm$url$Url$Parser$Internal$Parser(
			function (dict) {
				return func(
					a(dict));
			});
	});
var elm$url$Url$Parser$Query$custom = F2(
	function (key, func) {
		return elm$url$Url$Parser$Internal$Parser(
			function (dict) {
				return func(
					A2(
						elm$core$Maybe$withDefault,
						_List_Nil,
						A2(elm$core$Dict$get, key, dict)));
			});
	});
var elm$url$Url$Parser$Query$string = function (key) {
	return A2(
		elm$url$Url$Parser$Query$custom,
		key,
		function (stringList) {
			if (stringList.b && (!stringList.b.b)) {
				var str = stringList.a;
				return elm$core$Maybe$Just(str);
			} else {
				return elm$core$Maybe$Nothing;
			}
		});
};
var author$project$Routes$Ls$searchQueryFromUrl = function (url) {
	return A2(
		elm$core$Maybe$withDefault,
		'',
		A2(
			elm$url$Url$Parser$parse,
			elm$url$Url$Parser$query(
				A2(
					elm$url$Url$Parser$Query$map,
					elm$core$Maybe$withDefault(''),
					elm$url$Url$Parser$Query$string('filter'))),
			_Utils_update(
				url,
				{path: ''})));
};
var author$project$Routes$Ls$doListQueryFromUrl = function (url) {
	var path = author$project$Util$urlToPath(url);
	var filter = author$project$Routes$Ls$searchQueryFromUrl(url);
	return A3(author$project$Commands$doListQuery, author$project$Routes$Ls$GotResponse, path, filter);
};
var rundis$elm_bootstrap$Bootstrap$Alert$Shown = {$: 'Shown'};
var rundis$elm_bootstrap$Bootstrap$Alert$shown = rundis$elm_bootstrap$Bootstrap$Alert$Shown;
var rundis$elm_bootstrap$Bootstrap$Modal$Hide = {$: 'Hide'};
var rundis$elm_bootstrap$Bootstrap$Modal$hidden = rundis$elm_bootstrap$Bootstrap$Modal$Hide;
var author$project$Modals$History$newModel = function (rights) {
	return {alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, history: elm$core$Maybe$Nothing, lastPath: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, rights: rights};
};
var author$project$Modals$Mkdir$Ready = {$: 'Ready'};
var author$project$Modals$Mkdir$newModel = {alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, inputName: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Mkdir$Ready};
var author$project$Modals$MoveCopy$Copy = {$: 'Copy'};
var author$project$Modals$MoveCopy$Loading = {$: 'Loading'};
var author$project$Modals$MoveCopy$newCopyModel = {action: author$project$Modals$MoveCopy$Copy, alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, destPath: '', filter: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, sourcePath: '', state: author$project$Modals$MoveCopy$Loading};
var author$project$Modals$MoveCopy$Move = {$: 'Move'};
var author$project$Modals$MoveCopy$newMoveModel = {action: author$project$Modals$MoveCopy$Move, alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, destPath: '', filter: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, sourcePath: '', state: author$project$Modals$MoveCopy$Loading};
var author$project$Modals$Remove$Ready = {$: 'Ready'};
var author$project$Modals$Remove$newModel = {alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, selected: _List_Nil, state: author$project$Modals$Remove$Ready};
var author$project$Modals$Rename$Ready = {$: 'Ready'};
var author$project$Modals$Rename$newModel = {alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, currPath: '', inputName: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Rename$Ready};
var author$project$Modals$Share$newModel = {modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, paths: _List_Nil};
var author$project$Modals$Upload$newModel = {failed: _List_Nil, success: _List_Nil, uploads: elm$core$Dict$empty};
var author$project$Routes$Ls$Loading = {$: 'Loading'};
var author$project$Routes$Ls$newModel = F3(
	function (key, url, rights) {
		return {
			alert: rundis$elm_bootstrap$Bootstrap$Alert$closed,
			copyState: author$project$Modals$MoveCopy$newCopyModel,
			currError: '',
			historyState: author$project$Modals$History$newModel(rights),
			key: key,
			mkdirState: author$project$Modals$Mkdir$newModel,
			moveState: author$project$Modals$MoveCopy$newMoveModel,
			removeState: author$project$Modals$Remove$newModel,
			renameState: author$project$Modals$Rename$newModel,
			rights: rights,
			shareState: author$project$Modals$Share$newModel,
			state: author$project$Routes$Ls$Loading,
			uploadState: author$project$Modals$Upload$newModel,
			url: url,
			zone: elm$time$Time$utc
		};
	});
var author$project$Commands$Identity = F2(
	function (name, fingerprint) {
		return {fingerprint: fingerprint, name: name};
	});
var author$project$Commands$SelfResponse = F2(
	function (self, defaultConflictStrategy) {
		return {defaultConflictStrategy: defaultConflictStrategy, self: self};
	});
var author$project$Commands$emptySelf = A2(
	author$project$Commands$SelfResponse,
	A2(author$project$Commands$Identity, '', ''),
	'marker');
var author$project$Modals$RemoteAdd$Ready = {$: 'Ready'};
var author$project$Modals$RemoteAdd$newModelWithState = function (state) {
	return {acceptPush: false, alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, conflictDropdown: rundis$elm_bootstrap$Bootstrap$Dropdown$initialState, conflictStrategy: '', doAutoUdate: false, fingerprint: '', modal: state, name: '', state: author$project$Modals$RemoteAdd$Ready};
};
var author$project$Modals$RemoteAdd$newModel = author$project$Modals$RemoteAdd$newModelWithState(rundis$elm_bootstrap$Bootstrap$Modal$hidden);
var author$project$Commands$emptyRemote = {
	acceptAutoUpdates: false,
	acceptPush: false,
	conflictStrategy: '',
	fingerprint: '',
	folders: _List_Nil,
	isAuthenticated: false,
	isOnline: false,
	lastSeen: elm$time$Time$millisToPosix(0),
	name: ''
};
var author$project$Modals$RemoteFolders$Ready = {$: 'Ready'};
var author$project$Modals$RemoteFolders$newModelWithState = F2(
	function (state, remote) {
		return {alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, allDirs: _List_Nil, conflictDropdowns: elm$core$Dict$empty, filter: '', modal: state, remote: remote, state: author$project$Modals$RemoteFolders$Ready};
	});
var author$project$Modals$RemoteFolders$newModel = A2(author$project$Modals$RemoteFolders$newModelWithState, rundis$elm_bootstrap$Bootstrap$Modal$hidden, author$project$Commands$emptyRemote);
var author$project$Modals$RemoteRemove$Ready = {$: 'Ready'};
var author$project$Modals$RemoteRemove$newModelWithState = F2(
	function (name, state) {
		return {alert: rundis$elm_bootstrap$Bootstrap$Alert$shown, modal: state, name: name, state: author$project$Modals$RemoteRemove$Ready};
	});
var author$project$Modals$RemoteRemove$newModel = A2(author$project$Modals$RemoteRemove$newModelWithState, '', rundis$elm_bootstrap$Bootstrap$Modal$hidden);
var author$project$Routes$Remotes$Loading = {$: 'Loading'};
var author$project$Routes$Remotes$newModel = F3(
	function (key, zone, rights) {
		return {actionDropdowns: elm$core$Dict$empty, alert: author$project$Util$defaultAlertState, conflictDropdowns: elm$core$Dict$empty, key: key, remoteAddState: author$project$Modals$RemoteAdd$newModel, remoteFoldersState: author$project$Modals$RemoteFolders$newModel, remoteRemoveState: author$project$Modals$RemoteRemove$newModel, rights: rights, self: author$project$Commands$emptySelf, state: author$project$Routes$Remotes$Loading, zone: zone};
	});
var author$project$Commands$Remote = F9(
	function (name, folders, fingerprint, acceptAutoUpdates, isOnline, isAuthenticated, lastSeen, acceptPush, conflictStrategy) {
		return {acceptAutoUpdates: acceptAutoUpdates, acceptPush: acceptPush, conflictStrategy: conflictStrategy, fingerprint: fingerprint, folders: folders, isAuthenticated: isAuthenticated, isOnline: isOnline, lastSeen: lastSeen, name: name};
	});
var author$project$Commands$Folder = F3(
	function (folder, readOnly, conflictStrategy) {
		return {conflictStrategy: conflictStrategy, folder: folder, readOnly: readOnly};
	});
var author$project$Commands$decodeFolder = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'conflict_strategy',
	elm$json$Json$Decode$string,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'read_only',
		elm$json$Json$Decode$bool,
		A3(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
			'folder',
			elm$json$Json$Decode$string,
			elm$json$Json$Decode$succeed(author$project$Commands$Folder))));
var elm$json$Json$Decode$fail = _Json_fail;
var elm$regex$Regex$Match = F4(
	function (match, index, number, submatches) {
		return {index: index, match: match, number: number, submatches: submatches};
	});
var elm$regex$Regex$findAtMost = _Regex_findAtMost;
var elm$regex$Regex$fromStringWith = _Regex_fromStringWith;
var elm$regex$Regex$fromString = function (string) {
	return A2(
		elm$regex$Regex$fromStringWith,
		{caseInsensitive: false, multiline: false},
		string);
};
var elm$regex$Regex$never = _Regex_never;
var jweir$elm_iso8601$ISO8601$iso8601Regex = A2(
	elm$regex$Regex$findAtMost,
	1,
	A2(
		elm$core$Maybe$withDefault,
		elm$regex$Regex$never,
		elm$regex$Regex$fromString('(\\d{4})-?' + ('(\\d{2})?-?' + ('(\\d{2})?' + ('T?' + ('(\\d{2})?:?' + ('(\\d{2})?:?' + ('(\\d{2})?' + ('([.,]\\d{1,})?' + ('(Z|[+-]\\d{2}:?\\d{2})?' + '(.*)?')))))))))));
var elm$core$Basics$round = _Basics_round;
var elm$core$String$toFloat = _String_toFloat;
var elm$regex$Regex$replaceAtMost = _Regex_replaceAtMost;
var jweir$elm_iso8601$ISO8601$parseMilliseconds = function (msString) {
	if (msString.$ === 'Nothing') {
		return 0;
	} else {
		var s = msString.a;
		var decimalStr = A4(
			elm$regex$Regex$replaceAtMost,
			1,
			A2(
				elm$core$Maybe$withDefault,
				elm$regex$Regex$never,
				elm$regex$Regex$fromString('[,.]')),
			function (_n1) {
				return '0.';
			},
			s);
		var decimal = A2(
			elm$core$Maybe$withDefault,
			0.0,
			elm$core$String$toFloat(decimalStr));
		return elm$core$Basics$round(1000 * decimal);
	}
};
var jweir$elm_iso8601$ISO8601$Extras$toInt = function (str) {
	return A2(
		elm$core$Maybe$withDefault,
		0,
		elm$core$String$toInt(str));
};
var jweir$elm_iso8601$ISO8601$parseOffset = function (timeString) {
	var setHour = F2(
		function (modifier, hour_) {
			switch (modifier) {
				case '+':
					return hour_;
				case '-':
					return _Utils_ap(modifier, hour_);
				default:
					return hour_;
			}
		});
	var re = A2(
		elm$core$Maybe$withDefault,
		elm$regex$Regex$never,
		elm$regex$Regex$fromString('(Z|([+-]\\d{2}:?\\d{2}))?'));
	var match = A3(
		elm$regex$Regex$findAtMost,
		1,
		A2(
			elm$core$Maybe$withDefault,
			elm$regex$Regex$never,
			elm$regex$Regex$fromString('([-+])(\\d\\d):?(\\d\\d)')),
		A2(elm$core$Maybe$withDefault, '', timeString));
	var parts = A2(
		elm$core$List$map,
		function ($) {
			return $.submatches;
		},
		match);
	_n0$2:
	while (true) {
		if ((((parts.b && parts.a.b) && (parts.a.a.$ === 'Just')) && parts.a.b.b) && (parts.a.b.a.$ === 'Just')) {
			if (parts.a.b.b.b) {
				if (((parts.a.b.b.a.$ === 'Just') && (!parts.a.b.b.b.b)) && (!parts.b.b)) {
					var _n1 = parts.a;
					var modifier = _n1.a.a;
					var _n2 = _n1.b;
					var hour_ = _n2.a.a;
					var _n3 = _n2.b;
					var minute_ = _n3.a.a;
					return _Utils_Tuple2(
						jweir$elm_iso8601$ISO8601$Extras$toInt(
							A2(setHour, modifier, hour_)),
						jweir$elm_iso8601$ISO8601$Extras$toInt(minute_));
				} else {
					break _n0$2;
				}
			} else {
				if (!parts.b.b) {
					var _n4 = parts.a;
					var modifier = _n4.a.a;
					var _n5 = _n4.b;
					var hour_ = _n5.a.a;
					return _Utils_Tuple2(
						jweir$elm_iso8601$ISO8601$Extras$toInt(
							A2(setHour, modifier, hour_)),
						0);
				} else {
					break _n0$2;
				}
			}
		} else {
			break _n0$2;
		}
	}
	return _Utils_Tuple2(0, 0);
};
var jweir$elm_iso8601$ISO8601$validateHour = function (time) {
	var s = time.second;
	var m = time.minute;
	var h = time.hour;
	return ((h === 24) && ((m + s) > 0)) ? elm$core$Result$Err('hour is out of range') : (((h < 0) || (h > 24)) ? elm$core$Result$Err('hour is out of range') : (((m < 0) || (m > 59)) ? elm$core$Result$Err('minute is out of range') : (((s < 0) || (s > 59)) ? elm$core$Result$Err('second is out of range') : elm$core$Result$Ok(time))));
};
var elm$core$Bitwise$shiftRightZfBy = _Bitwise_shiftRightZfBy;
var elm$core$Array$bitMask = 4294967295 >>> (32 - elm$core$Array$shiftStep);
var elm$core$Bitwise$and = _Bitwise_and;
var elm$core$Elm$JsArray$unsafeGet = _JsArray_unsafeGet;
var elm$core$Array$getHelp = F3(
	function (shift, index, tree) {
		getHelp:
		while (true) {
			var pos = elm$core$Array$bitMask & (index >>> shift);
			var _n0 = A2(elm$core$Elm$JsArray$unsafeGet, pos, tree);
			if (_n0.$ === 'SubTree') {
				var subTree = _n0.a;
				var $temp$shift = shift - elm$core$Array$shiftStep,
					$temp$index = index,
					$temp$tree = subTree;
				shift = $temp$shift;
				index = $temp$index;
				tree = $temp$tree;
				continue getHelp;
			} else {
				var values = _n0.a;
				return A2(elm$core$Elm$JsArray$unsafeGet, elm$core$Array$bitMask & index, values);
			}
		}
	});
var elm$core$Bitwise$shiftLeftBy = _Bitwise_shiftLeftBy;
var elm$core$Array$tailIndex = function (len) {
	return (len >>> 5) << 5;
};
var elm$core$Basics$ge = _Utils_ge;
var elm$core$Array$get = F2(
	function (index, _n0) {
		var len = _n0.a;
		var startShift = _n0.b;
		var tree = _n0.c;
		var tail = _n0.d;
		return ((index < 0) || (_Utils_cmp(index, len) > -1)) ? elm$core$Maybe$Nothing : ((_Utils_cmp(
			index,
			elm$core$Array$tailIndex(len)) > -1) ? elm$core$Maybe$Just(
			A2(elm$core$Elm$JsArray$unsafeGet, elm$core$Array$bitMask & index, tail)) : elm$core$Maybe$Just(
			A3(elm$core$Array$getHelp, startShift, index, tree)));
	});
var elm$core$Array$fromListHelp = F3(
	function (list, nodeList, nodeListSize) {
		fromListHelp:
		while (true) {
			var _n0 = A2(elm$core$Elm$JsArray$initializeFromList, elm$core$Array$branchFactor, list);
			var jsArray = _n0.a;
			var remainingItems = _n0.b;
			if (_Utils_cmp(
				elm$core$Elm$JsArray$length(jsArray),
				elm$core$Array$branchFactor) < 0) {
				return A2(
					elm$core$Array$builderToArray,
					true,
					{nodeList: nodeList, nodeListSize: nodeListSize, tail: jsArray});
			} else {
				var $temp$list = remainingItems,
					$temp$nodeList = A2(
					elm$core$List$cons,
					elm$core$Array$Leaf(jsArray),
					nodeList),
					$temp$nodeListSize = nodeListSize + 1;
				list = $temp$list;
				nodeList = $temp$nodeList;
				nodeListSize = $temp$nodeListSize;
				continue fromListHelp;
			}
		}
	});
var elm$core$Array$fromList = function (list) {
	if (!list.b) {
		return elm$core$Array$empty;
	} else {
		return A3(elm$core$Array$fromListHelp, list, _List_Nil, 0);
	}
};
var jweir$elm_iso8601$ISO8601$Extras$calendar = elm$core$Array$fromList(
	_List_fromArray(
		[
			_Utils_Tuple3('January', 31, 31),
			_Utils_Tuple3('February', 28, 29),
			_Utils_Tuple3('March', 31, 31),
			_Utils_Tuple3('April', 30, 30),
			_Utils_Tuple3('May', 31, 31),
			_Utils_Tuple3('June', 30, 30),
			_Utils_Tuple3('July', 31, 31),
			_Utils_Tuple3('August', 31, 31),
			_Utils_Tuple3('September', 30, 30),
			_Utils_Tuple3('October', 31, 31),
			_Utils_Tuple3('November', 30, 30),
			_Utils_Tuple3('December', 31, 31)
		]));
var elm$core$Basics$modBy = _Basics_modBy;
var jweir$elm_iso8601$ISO8601$Extras$isLeapYear = function (year) {
	var c = !A2(elm$core$Basics$modBy, 400, year);
	var b = !A2(elm$core$Basics$modBy, 100, year);
	var a = !A2(elm$core$Basics$modBy, 4, year);
	var _n0 = _List_fromArray(
		[a, b, c]);
	_n0$3:
	while (true) {
		if ((_n0.b && _n0.a) && _n0.b.b) {
			if (!_n0.b.a) {
				if (_n0.b.b.b && (!_n0.b.b.b.b)) {
					var _n3 = _n0.b;
					var _n4 = _n3.b;
					return true;
				} else {
					break _n0$3;
				}
			} else {
				if (_n0.b.b.b) {
					if (_n0.b.b.a) {
						if (!_n0.b.b.b.b) {
							var _n1 = _n0.b;
							var _n2 = _n1.b;
							return true;
						} else {
							break _n0$3;
						}
					} else {
						if (!_n0.b.b.b.b) {
							var _n5 = _n0.b;
							var _n6 = _n5.b;
							return false;
						} else {
							break _n0$3;
						}
					}
				} else {
					break _n0$3;
				}
			}
		} else {
			break _n0$3;
		}
	}
	return false;
};
var jweir$elm_iso8601$ISO8601$Extras$daysInMonth = F2(
	function (year, monthInt) {
		var calMonth = A2(elm$core$Array$get, monthInt - 1, jweir$elm_iso8601$ISO8601$Extras$calendar);
		if (calMonth.$ === 'Just') {
			var _n1 = calMonth.a;
			var days = _n1.b;
			var leapDays = _n1.c;
			return jweir$elm_iso8601$ISO8601$Extras$isLeapYear(year) ? leapDays : days;
		} else {
			return 0;
		}
	});
var jweir$elm_iso8601$ISO8601$validateTime = function (time) {
	var maxDays = jweir$elm_iso8601$ISO8601$Extras$daysInMonth;
	return ((time.month < 1) || (time.month > 12)) ? elm$core$Result$Err('month is out of range') : (((time.day < 1) || (_Utils_cmp(
		time.day,
		A2(jweir$elm_iso8601$ISO8601$Extras$daysInMonth, time.year, time.month)) > 0)) ? elm$core$Result$Err('day is out of range') : jweir$elm_iso8601$ISO8601$validateHour(time));
};
var jweir$elm_iso8601$ISO8601$fromString = function (s) {
	var unwrap = F2(
		function (x, d) {
			return jweir$elm_iso8601$ISO8601$Extras$toInt(
				A2(elm$core$Maybe$withDefault, d, x));
		});
	var parts = A2(
		elm$core$List$map,
		function ($) {
			return $.submatches;
		},
		jweir$elm_iso8601$ISO8601$iso8601Regex(s));
	if (((((((((((parts.b && parts.a.b) && parts.a.b.b) && parts.a.b.b.b) && parts.a.b.b.b.b) && parts.a.b.b.b.b.b) && parts.a.b.b.b.b.b.b) && parts.a.b.b.b.b.b.b.b) && parts.a.b.b.b.b.b.b.b.b) && parts.a.b.b.b.b.b.b.b.b.b) && (!parts.a.b.b.b.b.b.b.b.b.b.b)) && (!parts.b.b)) {
		var _n1 = parts.a;
		var y = _n1.a;
		var _n2 = _n1.b;
		var mon = _n2.a;
		var _n3 = _n2.b;
		var d = _n3.a;
		var _n4 = _n3.b;
		var h = _n4.a;
		var _n5 = _n4.b;
		var min = _n5.a;
		var _n6 = _n5.b;
		var sec = _n6.a;
		var _n7 = _n6.b;
		var mil = _n7.a;
		var _n8 = _n7.b;
		var off = _n8.a;
		var _n9 = _n8.b;
		var invalid = _n9.a;
		if (invalid.$ === 'Just') {
			return elm$core$Result$Err('unexpected text');
		} else {
			return jweir$elm_iso8601$ISO8601$validateTime(
				{
					day: A2(unwrap, d, '1'),
					hour: A2(unwrap, h, '0'),
					millisecond: jweir$elm_iso8601$ISO8601$parseMilliseconds(mil),
					minute: A2(unwrap, min, '0'),
					month: A2(unwrap, mon, '1'),
					offset: jweir$elm_iso8601$ISO8601$parseOffset(off),
					second: A2(unwrap, sec, '0'),
					year: A2(unwrap, y, '0')
				});
		}
	} else {
		return elm$core$Result$Err('Unable to parse time');
	}
};
var elm$core$Basics$negate = function (n) {
	return -n;
};
var elm$core$List$sum = function (numbers) {
	return A3(elm$core$List$foldl, elm$core$Basics$add, 0, numbers);
};
var jweir$elm_iso8601$ISO8601$ims = 1;
var jweir$elm_iso8601$ISO8601$isec = jweir$elm_iso8601$ISO8601$ims * 1000;
var jweir$elm_iso8601$ISO8601$imin = jweir$elm_iso8601$ISO8601$isec * 60;
var jweir$elm_iso8601$ISO8601$ihour = jweir$elm_iso8601$ISO8601$imin * 60;
var jweir$elm_iso8601$ISO8601$iday = jweir$elm_iso8601$ISO8601$ihour * 24;
var jweir$elm_iso8601$ISO8601$offsetToTime = function (time) {
	var _n0 = time.offset;
	var m = _n0.a;
	var s = _n0.b;
	return (jweir$elm_iso8601$ISO8601$ihour * m) + (jweir$elm_iso8601$ISO8601$imin * s);
};
var jweir$elm_iso8601$ISO8601$Extras$daysInYear = function (year) {
	return jweir$elm_iso8601$ISO8601$Extras$isLeapYear(year) ? 366 : 365;
};
var jweir$elm_iso8601$ISO8601$toTime = function (time) {
	var _n0 = time.year >= 1970;
	if (!_n0) {
		var years = A2(
			elm$core$List$map,
			jweir$elm_iso8601$ISO8601$Extras$daysInYear,
			A2(elm$core$List$range, time.year + 1, 1970 - 1));
		var totalDays = elm$core$List$sum(
			A2(
				elm$core$List$map,
				jweir$elm_iso8601$ISO8601$Extras$daysInMonth(time.year),
				A2(elm$core$List$range, 1, time.month)));
		var tots = _List_fromArray(
			[
				jweir$elm_iso8601$ISO8601$iday * elm$core$List$sum(years),
				jweir$elm_iso8601$ISO8601$iday * (jweir$elm_iso8601$ISO8601$Extras$daysInYear(time.year) - totalDays),
				jweir$elm_iso8601$ISO8601$iday * (A2(jweir$elm_iso8601$ISO8601$Extras$daysInMonth, time.year, time.month) - time.day),
				(jweir$elm_iso8601$ISO8601$iday - jweir$elm_iso8601$ISO8601$ihour) - (jweir$elm_iso8601$ISO8601$ihour * time.hour),
				(jweir$elm_iso8601$ISO8601$ihour - jweir$elm_iso8601$ISO8601$imin) - (jweir$elm_iso8601$ISO8601$imin * time.minute),
				jweir$elm_iso8601$ISO8601$imin - (jweir$elm_iso8601$ISO8601$isec * time.second),
				jweir$elm_iso8601$ISO8601$offsetToTime(time)
			]);
		return 0 - (elm$core$List$sum(tots) - time.millisecond);
	} else {
		var years = A2(
			elm$core$List$map,
			jweir$elm_iso8601$ISO8601$Extras$daysInYear,
			A2(elm$core$List$range, 1970, time.year - 1));
		var months = A2(
			elm$core$List$map,
			jweir$elm_iso8601$ISO8601$Extras$daysInMonth(time.year),
			A2(elm$core$List$range, 1, time.month - 1));
		var tots = _List_fromArray(
			[
				jweir$elm_iso8601$ISO8601$iday * elm$core$List$sum(years),
				jweir$elm_iso8601$ISO8601$iday * elm$core$List$sum(months),
				jweir$elm_iso8601$ISO8601$iday * (time.day - 1),
				jweir$elm_iso8601$ISO8601$ihour * time.hour,
				jweir$elm_iso8601$ISO8601$imin * time.minute,
				jweir$elm_iso8601$ISO8601$isec * time.second,
				(-1) * jweir$elm_iso8601$ISO8601$offsetToTime(time)
			]);
		return elm$core$List$sum(tots) + time.millisecond;
	}
};
var jweir$elm_iso8601$ISO8601$toPosix = function (time) {
	return elm$time$Time$millisToPosix(
		jweir$elm_iso8601$ISO8601$toTime(time));
};
var author$project$Commands$iso8601ToPosix = A2(
	elm$json$Json$Decode$andThen,
	function (stamp) {
		var _n0 = jweir$elm_iso8601$ISO8601$fromString(stamp);
		if (_n0.$ === 'Ok') {
			var time = _n0.a;
			return elm$json$Json$Decode$succeed(
				jweir$elm_iso8601$ISO8601$toPosix(time));
		} else {
			var msg = _n0.a;
			return elm$json$Json$Decode$fail(msg);
		}
	},
	elm$json$Json$Decode$string);
var elm$json$Json$Decode$null = _Json_decodeNull;
var elm$json$Json$Decode$oneOf = _Json_oneOf;
var author$project$Commands$decodeRemote = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'conflict_strategy',
	elm$json$Json$Decode$string,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'accept_push',
		elm$json$Json$Decode$bool,
		A3(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
			'last_seen',
			author$project$Commands$iso8601ToPosix,
			A3(
				NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
				'is_authenticated',
				elm$json$Json$Decode$bool,
				A3(
					NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
					'is_online',
					elm$json$Json$Decode$bool,
					A3(
						NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
						'accept_auto_updates',
						elm$json$Json$Decode$bool,
						A3(
							NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
							'fingerprint',
							elm$json$Json$Decode$string,
							A3(
								NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
								'folders',
								elm$json$Json$Decode$oneOf(
									_List_fromArray(
										[
											elm$json$Json$Decode$list(author$project$Commands$decodeFolder),
											elm$json$Json$Decode$null(_List_Nil)
										])),
								A3(
									NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
									'name',
									elm$json$Json$Decode$string,
									elm$json$Json$Decode$succeed(author$project$Commands$Remote))))))))));
var author$project$Commands$decodeRemoteListResponse = A2(
	elm$json$Json$Decode$field,
	'remotes',
	elm$json$Json$Decode$list(author$project$Commands$decodeRemote));
var author$project$Commands$doRemoteList = function (toMsg) {
	return elm$http$Http$post(
		{
			body: elm$http$Http$emptyBody,
			expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeRemoteListResponse),
			url: '/api/v0/remotes/list'
		});
};
var author$project$Commands$decodeIdentity = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'fingerprint',
	elm$json$Json$Decode$string,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'name',
		elm$json$Json$Decode$string,
		elm$json$Json$Decode$succeed(author$project$Commands$Identity)));
var author$project$Commands$decodeSelfResponse = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'default_conflict_strategy',
	elm$json$Json$Decode$string,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'self',
		author$project$Commands$decodeIdentity,
		elm$json$Json$Decode$succeed(author$project$Commands$SelfResponse)));
var author$project$Commands$doSelfQuery = function (toMsg) {
	return elm$http$Http$post(
		{
			body: elm$http$Http$emptyBody,
			expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeSelfResponse),
			url: '/api/v0/remotes/self'
		});
};
var author$project$Routes$Remotes$GotRemoteListResponse = function (a) {
	return {$: 'GotRemoteListResponse', a: a};
};
var author$project$Routes$Remotes$GotSelfResponse = function (a) {
	return {$: 'GotSelfResponse', a: a};
};
var author$project$Routes$Remotes$reload = elm$core$Platform$Cmd$batch(
	_List_fromArray(
		[
			author$project$Commands$doRemoteList(author$project$Routes$Remotes$GotRemoteListResponse),
			author$project$Commands$doSelfQuery(author$project$Routes$Remotes$GotSelfResponse)
		]));
var elm$json$Json$Encode$null = _Json_encodeNull;
var author$project$Websocket$open = _Platform_outgoingPort(
	'open',
	function ($) {
		return elm$json$Json$Encode$null;
	});
var elm$core$Platform$Cmd$map = _Platform_map;
var author$project$Main$doInitAfterLogin = F5(
	function (model, loginName, rights, isAnon, anonIsAllowed) {
		var newViewState = {
			anonIsAllowed: anonIsAllowed,
			commitsState: A4(author$project$Routes$Commits$newModel, model.url, model.key, model.zone, rights),
			currentView: A2(author$project$Main$viewFromUrl, rights, model.url),
			deletedFilesState: A4(author$project$Routes$DeletedFiles$newModel, model.url, model.key, model.zone, rights),
			diffState: A3(author$project$Routes$Diff$newModel, model.key, model.url, model.zone),
			isAnon: isAnon,
			listState: A3(author$project$Routes$Ls$newModel, model.key, model.url, rights),
			loginName: loginName,
			remoteState: A3(author$project$Routes$Remotes$newModel, model.key, model.zone, rights),
			rights: rights
		};
		return _Utils_Tuple2(
			_Utils_update(
				model,
				{
					loginState: author$project$Main$LoginSuccess(newViewState)
				}),
			elm$core$Platform$Cmd$batch(
				_List_fromArray(
					[
						A2(
						elm$core$Platform$Cmd$map,
						author$project$Main$ListMsg,
						author$project$Routes$Ls$doListQueryFromUrl(model.url)),
						author$project$Websocket$open(_Utils_Tuple0),
						A2(
						elm$core$Platform$Cmd$map,
						author$project$Main$DeletedFilesMsg,
						author$project$Routes$DeletedFiles$reload(newViewState.deletedFilesState)),
						A2(
						elm$core$Platform$Cmd$map,
						author$project$Main$CommitsMsg,
						author$project$Routes$Commits$reload(newViewState.commitsState)),
						A2(elm$core$Platform$Cmd$map, author$project$Main$RemotesMsg, author$project$Routes$Remotes$reload),
						A2(
						elm$core$Platform$Cmd$map,
						author$project$Main$DiffMsg,
						A2(author$project$Routes$Diff$reload, newViewState.diffState, model.url))
					])));
	});
var author$project$Main$eventType = function (data) {
	var result = A2(
		elm$json$Json$Decode$decodeString,
		A2(elm$json$Json$Decode$field, 'data', elm$json$Json$Decode$string),
		data);
	if (result.$ === 'Ok') {
		var typ = result.a;
		return typ;
	} else {
		return 'failed';
	}
};
var author$project$Main$pingerMsgToBool = function (data) {
	var result = A2(
		elm$json$Json$Decode$decodeString,
		A2(elm$json$Json$Decode$field, 'isOnline', elm$json$Json$Decode$bool),
		data);
	if (result.$ === 'Ok') {
		var typ = result.a;
		return typ;
	} else {
		return false;
	}
};
var author$project$Main$withSubUpdate = F6(
	function (subMsg, subModel, model, msg, subUpdate, viewStateUpdate) {
		var _n0 = model.loginState;
		if (_n0.$ === 'LoginSuccess') {
			var viewState = _n0.a;
			var _n1 = A2(
				subUpdate,
				subMsg,
				subModel(viewState));
			var newSubModel = _n1.a;
			var newSubCmd = _n1.b;
			return _Utils_Tuple2(
				_Utils_update(
					model,
					{
						loginState: author$project$Main$LoginSuccess(
							A2(viewStateUpdate, viewState, newSubModel))
					}),
				A2(elm$core$Platform$Cmd$map, msg, newSubCmd));
		} else {
			return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
		}
	});
var author$project$Routes$Commits$reloadIfNeeded = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var commits = _n0.a;
		return (!elm$core$List$length(commits)) ? author$project$Routes$Commits$reload(model) : elm$core$Platform$Cmd$none;
	} else {
		return elm$core$Platform$Cmd$none;
	}
};
var author$project$Commands$ResetQuery = F3(
	function (path, revision, force) {
		return {force: force, path: path, revision: revision};
	});
var author$project$Commands$decodeResetQuery = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodeResetQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'path',
				elm$json$Json$Encode$string(q.path)),
				_Utils_Tuple2(
				'revision',
				elm$json$Json$Encode$string(q.revision))
			]));
};
var author$project$Commands$doReset = F3(
	function (toMsg, path, revision) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeResetQuery(
						A3(author$project$Commands$ResetQuery, path, revision, true))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeResetQuery),
				url: '/api/v0/reset'
			});
	});
var author$project$Routes$Commits$Failure = function (a) {
	return {$: 'Failure', a: a};
};
var author$project$Routes$Commits$GotResetResponse = function (a) {
	return {$: 'GotResetResponse', a: a};
};
var author$project$Routes$Commits$Success = function (a) {
	return {$: 'Success', a: a};
};
var author$project$Routes$Commits$toMap = function (commits) {
	return elm$core$Dict$fromList(
		A2(
			elm$core$List$map,
			function (c) {
				return _Utils_Tuple2(c.index, c);
			},
			commits));
};
var author$project$Routes$Commits$mergeCommits = F2(
	function (old, _new) {
		return elm$core$List$reverse(
			A2(
				elm$core$List$map,
				function (_n0) {
					var v = _n0.b;
					return v;
				},
				elm$core$Dict$toList(
					A2(
						elm$core$Dict$union,
						author$project$Routes$Commits$toMap(_new),
						author$project$Routes$Commits$toMap(old)))));
	});
var author$project$Routes$Commits$reloadWithoutFlush = F2(
	function (model, newOffset) {
		return A4(
			author$project$Commands$doLog,
			author$project$Routes$Commits$GotLogResponse(false),
			newOffset,
			author$project$Routes$Commits$loadLimit,
			model.filter);
	});
var andrewMacmurray$elm_delay$Delay$Second = {$: 'Second'};
var andrewMacmurray$elm_delay$Delay$Duration = F2(
	function (a, b) {
		return {$: 'Duration', a: a, b: b};
	});
var elm$core$Basics$always = F2(
	function (a, _n0) {
		return a;
	});
var elm$core$Process$sleep = _Process_sleep;
var andrewMacmurray$elm_delay$Delay$after_ = F2(
	function (time, msg) {
		return A2(
			elm$core$Task$perform,
			elm$core$Basics$always(msg),
			elm$core$Process$sleep(time));
	});
var andrewMacmurray$elm_delay$Delay$Minute = {$: 'Minute'};
var andrewMacmurray$elm_delay$Delay$toMillis = function (_n0) {
	var t = _n0.a;
	var u = _n0.b;
	switch (u.$) {
		case 'Millisecond':
			return t;
		case 'Second':
			return 1000 * t;
		case 'Minute':
			return andrewMacmurray$elm_delay$Delay$toMillis(
				A2(andrewMacmurray$elm_delay$Delay$Duration, 60 * t, andrewMacmurray$elm_delay$Delay$Second));
		default:
			return andrewMacmurray$elm_delay$Delay$toMillis(
				A2(andrewMacmurray$elm_delay$Delay$Duration, 60 * t, andrewMacmurray$elm_delay$Delay$Minute));
	}
};
var andrewMacmurray$elm_delay$Delay$after = F3(
	function (time, unit, msg) {
		return A2(
			andrewMacmurray$elm_delay$Delay$after_,
			andrewMacmurray$elm_delay$Delay$toMillis(
				A2(andrewMacmurray$elm_delay$Delay$Duration, time, unit)),
			msg);
	});
var author$project$Routes$Commits$AlertMsg = function (a) {
	return {$: 'AlertMsg', a: a};
};
var author$project$Util$AlertState = F3(
	function (message, typ, vis) {
		return {message: message, typ: typ, vis: vis};
	});
var author$project$Routes$Commits$showAlert = F4(
	function (model, duration, modalTyp, message) {
		var newAlert = A3(author$project$Util$AlertState, message, modalTyp, rundis$elm_bootstrap$Bootstrap$Alert$shown);
		return _Utils_Tuple2(
			_Utils_update(
				model,
				{alert: newAlert}),
			elm$core$Platform$Cmd$batch(
				_List_fromArray(
					[
						A3(
						andrewMacmurray$elm_delay$Delay$after,
						duration,
						andrewMacmurray$elm_delay$Delay$Second,
						author$project$Routes$Commits$AlertMsg(rundis$elm_bootstrap$Bootstrap$Alert$closed))
					])));
	});
var author$project$Scroll$percFloat = function (data) {
	return (data.scrollTop * 100) / (data.pageHeight - data.viewportHeight);
};
var author$project$Scroll$hasHitBottom = function (data) {
	return author$project$Scroll$percFloat(data) >= 95;
};
var author$project$Util$Danger = {$: 'Danger'};
var author$project$Util$Success = {$: 'Success'};
var author$project$Util$httpErrorToString = function (err) {
	switch (err.$) {
		case 'BadUrl':
			var msg = err.a;
			return 'Bad url: ' + msg;
		case 'Timeout':
			return 'Timeout';
		case 'NetworkError':
			return 'Network error';
		case 'BadStatus':
			var status = err.a;
			return 'Bad status: ' + elm$core$String$fromInt(status);
		default:
			var msg = err.a;
			return 'Could not decode body: ' + msg;
	}
};
var author$project$Routes$Commits$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'GotLogResponse':
				var doFlush = msg.a;
				var result = msg.b;
				if (result.$ === 'Ok') {
					var log = result.a;
					var _n2 = function () {
						if (doFlush) {
							return _Utils_Tuple2(_List_Nil, 0);
						} else {
							var _n3 = model.state;
							if (_n3.$ === 'Success') {
								var oldCommits = _n3.a;
								return _Utils_Tuple2(oldCommits, model.offset + author$project$Routes$Commits$loadLimit);
							} else {
								return _Utils_Tuple2(_List_Nil, model.offset);
							}
						}
					}();
					var prevCommits = _n2.a;
					var newOffset = _n2.b;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								haveStagedChanges: log.haveStagedChanges,
								offset: newOffset,
								state: author$project$Routes$Commits$Success(
									A2(author$project$Routes$Commits$mergeCommits, prevCommits, log.commits))
							}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Routes$Commits$Failure(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotResetResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return A4(author$project$Routes$Commits$showAlert, model, 5, author$project$Util$Success, 'Succesfully reset state.');
				} else {
					var err = result.a;
					return A4(
						author$project$Routes$Commits$showAlert,
						model,
						15,
						author$project$Util$Danger,
						'Failed to reset: ' + author$project$Util$httpErrorToString(err));
				}
			case 'CheckoutClicked':
				var hash = msg.a;
				return _Utils_Tuple2(
					model,
					A3(author$project$Commands$doReset, author$project$Routes$Commits$GotResetResponse, '/', hash));
			case 'SearchInput':
				var filter = msg.a;
				var upModel = _Utils_update(
					model,
					{filter: filter});
				return _Utils_Tuple2(
					upModel,
					author$project$Routes$Commits$reload(upModel));
			case 'OnScroll':
				var data = msg.a;
				return A2(elm$core$String$startsWith, '/log', model.url.path) ? (author$project$Scroll$hasHitBottom(data) ? _Utils_Tuple2(
					model,
					A2(author$project$Routes$Commits$reloadWithoutFlush, model, model.offset + author$project$Routes$Commits$loadLimit)) : _Utils_Tuple2(model, elm$core$Platform$Cmd$none)) : _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
			default:
				var vis = msg.a;
				var newAlert = A3(author$project$Util$AlertState, model.alert.message, model.alert.typ, vis);
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: newAlert}),
					elm$core$Platform$Cmd$none);
		}
	});
var author$project$Routes$Commits$updateUrl = F2(
	function (model, url) {
		return _Utils_update(
			model,
			{url: url});
	});
var author$project$Routes$DeletedFiles$reloadIfNeeded = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var commits = _n0.a;
		return (!elm$core$List$length(commits)) ? author$project$Routes$DeletedFiles$reload(model) : elm$core$Platform$Cmd$none;
	} else {
		return elm$core$Platform$Cmd$none;
	}
};
var author$project$Commands$UndeleteQuery = function (path) {
	return {path: path};
};
var author$project$Commands$decodeUndeleteResponse = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodeUndeleteQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'path',
				elm$json$Json$Encode$string(q.path))
			]));
};
var author$project$Commands$doUndelete = F2(
	function (toMsg, path) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeUndeleteQuery(
						author$project$Commands$UndeleteQuery(path))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeUndeleteResponse),
				url: '/api/v0/undelete'
			});
	});
var author$project$Routes$DeletedFiles$Failure = function (a) {
	return {$: 'Failure', a: a};
};
var author$project$Routes$DeletedFiles$GotUndeleteResponse = function (a) {
	return {$: 'GotUndeleteResponse', a: a};
};
var author$project$Routes$DeletedFiles$Success = function (a) {
	return {$: 'Success', a: a};
};
var author$project$Routes$DeletedFiles$sortEntries = F2(
	function (a, b) {
		var inv = function (v) {
			return v ? 0 : 1;
		};
		var _n0 = A2(
			elm$core$Basics$compare,
			inv(a.isDir),
			inv(b.isDir));
		if (_n0.$ === 'EQ') {
			return A2(elm$core$Basics$compare, a.path, b.path);
		} else {
			var other = _n0;
			return other;
		}
	});
var author$project$Routes$DeletedFiles$toMap = function (entries) {
	return elm$core$Dict$fromList(
		A2(
			elm$core$List$map,
			function (e) {
				return _Utils_Tuple2(e.path, e);
			},
			entries));
};
var elm$core$List$sortWith = _List_sortWith;
var author$project$Routes$DeletedFiles$mergeEntries = F2(
	function (old, _new) {
		return A2(
			elm$core$List$sortWith,
			author$project$Routes$DeletedFiles$sortEntries,
			A2(
				elm$core$List$map,
				function (_n0) {
					var v = _n0.b;
					return v;
				},
				elm$core$Dict$toList(
					A2(
						elm$core$Dict$union,
						author$project$Routes$DeletedFiles$toMap(_new),
						author$project$Routes$DeletedFiles$toMap(old)))));
	});
var author$project$Routes$DeletedFiles$reloadWithoutFlush = F2(
	function (model, newOffset) {
		return A4(
			author$project$Commands$doDeletedFiles,
			author$project$Routes$DeletedFiles$GotDeletedPathsResponse(false),
			newOffset,
			author$project$Routes$DeletedFiles$loadLimit,
			model.filter);
	});
var author$project$Routes$DeletedFiles$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'GotDeletedPathsResponse':
				var doFlush = msg.a;
				var result = msg.b;
				if (result.$ === 'Ok') {
					var entries = result.a;
					var _n2 = function () {
						if (doFlush) {
							return _Utils_Tuple2(_List_Nil, 0);
						} else {
							var _n3 = model.state;
							if (_n3.$ === 'Success') {
								var oldEntries = _n3.a;
								return _Utils_Tuple2(oldEntries, model.offset + author$project$Routes$DeletedFiles$loadLimit);
							} else {
								return _Utils_Tuple2(_List_Nil, model.offset);
							}
						}
					}();
					var prevEntries = _n2.a;
					var newOffset = _n2.b;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								offset: newOffset,
								state: author$project$Routes$DeletedFiles$Success(
									A2(author$project$Routes$DeletedFiles$mergeEntries, prevEntries, entries))
							}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Routes$DeletedFiles$Failure(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'UndeleteClicked':
				var path = msg.a;
				return _Utils_Tuple2(
					model,
					A2(author$project$Commands$doUndelete, author$project$Routes$DeletedFiles$GotUndeleteResponse, path));
			case 'SearchInput':
				var filter = msg.a;
				var upModel = _Utils_update(
					model,
					{filter: filter});
				return _Utils_Tuple2(
					upModel,
					author$project$Routes$DeletedFiles$reload(upModel));
			case 'GotUndeleteResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var newAlert = A3(author$project$Util$AlertState, 'Succcesfully undeleted one item.', author$project$Util$Success, rundis$elm_bootstrap$Bootstrap$Alert$shown);
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{alert: newAlert}),
						elm$core$Platform$Cmd$batch(
							_List_fromArray(
								[
									author$project$Routes$DeletedFiles$reload(model),
									A3(
									andrewMacmurray$elm_delay$Delay$after,
									5,
									andrewMacmurray$elm_delay$Delay$Second,
									author$project$Routes$DeletedFiles$AlertMsg(rundis$elm_bootstrap$Bootstrap$Alert$closed))
								])));
				} else {
					var err = result.a;
					var newAlert = A3(
						author$project$Util$AlertState,
						'Failed to undelete: ' + author$project$Util$httpErrorToString(err),
						author$project$Util$Danger,
						rundis$elm_bootstrap$Bootstrap$Alert$shown);
					return _Utils_Tuple2(
						model,
						elm$core$Platform$Cmd$batch(
							_List_fromArray(
								[
									author$project$Routes$DeletedFiles$reload(model),
									A3(
									andrewMacmurray$elm_delay$Delay$after,
									15,
									andrewMacmurray$elm_delay$Delay$Second,
									author$project$Routes$DeletedFiles$AlertMsg(rundis$elm_bootstrap$Bootstrap$Alert$closed))
								])));
				}
			case 'AlertMsg':
				var vis = msg.a;
				var newAlert = A3(author$project$Util$AlertState, model.alert.message, model.alert.typ, vis);
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: newAlert}),
					elm$core$Platform$Cmd$none);
			default:
				var data = msg.a;
				return A2(elm$core$String$startsWith, '/deleted', model.url.path) ? (author$project$Scroll$hasHitBottom(data) ? _Utils_Tuple2(
					model,
					A2(author$project$Routes$DeletedFiles$reloadWithoutFlush, model, model.offset + author$project$Routes$DeletedFiles$loadLimit)) : _Utils_Tuple2(model, elm$core$Platform$Cmd$none)) : _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
		}
	});
var author$project$Routes$DeletedFiles$updateUrl = F2(
	function (model, url) {
		return _Utils_update(
			model,
			{url: url});
	});
var author$project$Routes$Diff$Finished = function (a) {
	return {$: 'Finished', a: a};
};
var elm$browser$Browser$Navigation$back = F2(
	function (key, n) {
		return A2(_Browser_go, key, -n);
	});
var author$project$Routes$Diff$update = F2(
	function (msg, model) {
		if (msg.$ === 'GotResponse') {
			var result = msg.a;
			if (result.$ === 'Ok') {
				var diff = result.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							state: author$project$Routes$Diff$Finished(
								elm$core$Result$Ok(diff))
						}),
					elm$core$Platform$Cmd$none);
			} else {
				var err = result.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							state: author$project$Routes$Diff$Finished(
								elm$core$Result$Err(
									author$project$Util$httpErrorToString(err)))
						}),
					elm$core$Platform$Cmd$none);
			}
		} else {
			return _Utils_Tuple2(
				model,
				A2(elm$browser$Browser$Navigation$back, model.key, 1));
		}
	});
var author$project$Routes$Diff$updateUrl = F2(
	function (model, url) {
		return _Utils_update(
			model,
			{url: url});
	});
var author$project$Routes$Ls$changeTimeZone = F2(
	function (zone, model) {
		return _Utils_update(
			model,
			{zone: zone});
	});
var author$project$Routes$Ls$changeUrl = F2(
	function (url, model) {
		return _Utils_update(
			model,
			{url: url});
	});
var author$project$Commands$PinQuery = F2(
	function (path, revision) {
		return {path: path, revision: revision};
	});
var author$project$Commands$decodePinResponse = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodePinQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'path',
				elm$json$Json$Encode$string(q.path)),
				_Utils_Tuple2(
				'revision',
				elm$json$Json$Encode$string(q.revision))
			]));
};
var author$project$Commands$doPin = F3(
	function (toMsg, path, revision) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodePinQuery(
						A2(author$project$Commands$PinQuery, path, revision))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodePinResponse),
				url: '/api/v0/pin'
			});
	});
var author$project$Commands$RemoveQuery = function (paths) {
	return {paths: paths};
};
var author$project$Commands$decodeRemoveResponse = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var elm$json$Json$Encode$list = F2(
	function (func, entries) {
		return _Json_wrap(
			A3(
				elm$core$List$foldl,
				_Json_addEntry(func),
				_Json_emptyArray(_Utils_Tuple0),
				entries));
	});
var author$project$Commands$encodeRemoveQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'paths',
				A2(elm$json$Json$Encode$list, elm$json$Json$Encode$string, q.paths))
			]));
};
var author$project$Commands$doRemove = F2(
	function (toMsg, paths) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeRemoveQuery(
						author$project$Commands$RemoveQuery(paths))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeRemoveResponse),
				url: '/api/v0/remove'
			});
	});
var author$project$Commands$doUnpin = F3(
	function (toMsg, path, revision) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodePinQuery(
						A2(author$project$Commands$PinQuery, path, revision))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodePinResponse),
				url: '/api/v0/unpin'
			});
	});
var author$project$Commands$HistoryQuery = function (path) {
	return {path: path};
};
var author$project$Commands$HistoryEntry = F5(
	function (head, path, change, isPinned, isExplicit) {
		return {change: change, head: head, isExplicit: isExplicit, isPinned: isPinned, path: path};
	});
var author$project$Commands$decodeHistoryEntry = A3(
	NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
	'is_explicit',
	elm$json$Json$Decode$bool,
	A3(
		NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
		'is_pinned',
		elm$json$Json$Decode$bool,
		A3(
			NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
			'change',
			elm$json$Json$Decode$string,
			A3(
				NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
				'path',
				elm$json$Json$Decode$string,
				A3(
					NoRedInk$elm_json_decode_pipeline$Json$Decode$Pipeline$required,
					'head',
					author$project$Commands$decodeCommit,
					elm$json$Json$Decode$succeed(author$project$Commands$HistoryEntry))))));
var author$project$Commands$decodeHistory = A2(
	elm$json$Json$Decode$field,
	'entries',
	elm$json$Json$Decode$list(author$project$Commands$decodeHistoryEntry));
var author$project$Commands$encodeHistoryQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'path',
				elm$json$Json$Encode$string(q.path))
			]));
};
var author$project$Commands$doHistory = F2(
	function (toMsg, path) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeHistoryQuery(
						author$project$Commands$HistoryQuery(path))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeHistory),
				url: '/api/v0/history'
			});
	});
var author$project$Modals$History$GotHistoryResponse = F2(
	function (a, b) {
		return {$: 'GotHistoryResponse', a: a, b: b};
	});
var author$project$Modals$History$show = function (path) {
	return A2(
		author$project$Commands$doHistory,
		author$project$Modals$History$GotHistoryResponse(path),
		path);
};
var author$project$Modals$History$GotPinResponse = function (a) {
	return {$: 'GotPinResponse', a: a};
};
var author$project$Modals$History$GotResetResponse = function (a) {
	return {$: 'GotResetResponse', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Modal$Show = {$: 'Show'};
var rundis$elm_bootstrap$Bootstrap$Modal$shown = rundis$elm_bootstrap$Bootstrap$Modal$Show;
var author$project$Modals$History$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'GotHistoryResponse':
				var path = msg.a;
				var result = msg.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							history: elm$core$Maybe$Just(result),
							lastPath: path,
							modal: rundis$elm_bootstrap$Bootstrap$Modal$shown
						}),
					elm$core$Platform$Cmd$none);
			case 'ResetClicked':
				var path = msg.a;
				var revision = msg.b;
				return _Utils_Tuple2(
					model,
					A3(author$project$Commands$doReset, author$project$Modals$History$GotResetResponse, path, revision));
			case 'GotResetResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{history: elm$core$Maybe$Nothing, modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								history: elm$core$Maybe$Just(
									elm$core$Result$Err(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotPinResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						model,
						A2(
							author$project$Commands$doHistory,
							author$project$Modals$History$GotHistoryResponse(model.lastPath),
							model.lastPath));
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								history: elm$core$Maybe$Just(
									elm$core$Result$Err(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$shown}),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{history: elm$core$Maybe$Nothing, modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			case 'PinClicked':
				var path = msg.a;
				var revision = msg.b;
				var shouldBePinned = msg.c;
				return _Utils_Tuple2(
					model,
					shouldBePinned ? A3(author$project$Commands$doPin, author$project$Modals$History$GotPinResponse, path, revision) : A3(author$project$Commands$doUnpin, author$project$Modals$History$GotPinResponse, path, revision));
			default:
				var key = msg.a;
				if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden)) {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					if (key === 'Enter') {
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{history: elm$core$Maybe$Nothing, modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
							elm$core$Platform$Cmd$none);
					} else {
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
					}
				}
		}
	});
var author$project$Commands$MkdirQuery = function (path) {
	return {path: path};
};
var author$project$Commands$decodeMkdirResponse = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodeMkdirQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'path',
				elm$json$Json$Encode$string(q.path))
			]));
};
var author$project$Commands$doMkdir = F2(
	function (toMsg, path) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeMkdirQuery(
						author$project$Commands$MkdirQuery(path))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeMkdirResponse),
				url: '/api/v0/mkdir'
			});
	});
var author$project$Modals$Mkdir$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Modals$Mkdir$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var author$project$Modals$Mkdir$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'CreateDir':
				var path = msg.a;
				return _Utils_Tuple2(
					model,
					A2(author$project$Commands$doMkdir, author$project$Modals$Mkdir$GotResponse, path));
			case 'InputChanged':
				var inputName = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{inputName: inputName}),
					elm$core$Platform$Cmd$none);
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Mkdir$Ready}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$Mkdir$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{inputName: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$shown}),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Mkdir$Ready}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			default:
				var path = msg.a;
				var key = msg.b;
				if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden)) {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					if (key === 'Enter') {
						return _Utils_Tuple2(
							model,
							A2(author$project$Commands$doMkdir, author$project$Modals$Mkdir$GotResponse, path));
					} else {
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
					}
				}
		}
	});
var author$project$Commands$decodeAllDirsResponse = A2(
	elm$json$Json$Decode$field,
	'paths',
	elm$json$Json$Decode$list(elm$json$Json$Decode$string));
var author$project$Commands$doListAllDirs = function (toMsg) {
	return elm$http$Http$post(
		{
			body: elm$http$Http$emptyBody,
			expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeAllDirsResponse),
			url: '/api/v0/all-dirs'
		});
};
var author$project$Modals$MoveCopy$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Modals$MoveCopy$GotAllDirsResponse = function (a) {
	return {$: 'GotAllDirsResponse', a: a};
};
var author$project$Modals$MoveCopy$Ready = function (a) {
	return {$: 'Ready', a: a};
};
var author$project$Commands$CopyQuery = F2(
	function (sourcePath, destinationPath) {
		return {destinationPath: destinationPath, sourcePath: sourcePath};
	});
var author$project$Commands$decodeCopyResponse = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Util$prefixSlash = function (path) {
	return A2(elm$core$String$startsWith, '/', path) ? path : ('/' + path);
};
var author$project$Commands$encodeCopyQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'source',
				elm$json$Json$Encode$string(
					author$project$Util$prefixSlash(q.sourcePath))),
				_Utils_Tuple2(
				'destination',
				elm$json$Json$Encode$string(
					author$project$Util$prefixSlash(q.destinationPath)))
			]));
};
var author$project$Commands$doCopy = F3(
	function (toMsg, src, dst) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeCopyQuery(
						A2(author$project$Commands$CopyQuery, src, dst))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeCopyResponse),
				url: '/api/v0/copy'
			});
	});
var author$project$Commands$MoveQuery = F2(
	function (sourcePath, destinationPath) {
		return {destinationPath: destinationPath, sourcePath: sourcePath};
	});
var author$project$Commands$decodeMoveResponse = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodeMoveQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'source',
				elm$json$Json$Encode$string(
					author$project$Util$prefixSlash(q.sourcePath))),
				_Utils_Tuple2(
				'destination',
				elm$json$Json$Encode$string(
					author$project$Util$prefixSlash(q.destinationPath)))
			]));
};
var author$project$Commands$doMove = F3(
	function (toMsg, src, dst) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeMoveQuery(
						A2(author$project$Commands$MoveQuery, src, dst))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeMoveResponse),
				url: '/api/v0/move'
			});
	});
var author$project$Modals$MoveCopy$GotActionResponse = function (a) {
	return {$: 'GotActionResponse', a: a};
};
var author$project$Modals$MoveCopy$doAction = function (model) {
	var _n0 = model.action;
	if (_n0.$ === 'Move') {
		return A3(author$project$Commands$doMove, author$project$Modals$MoveCopy$GotActionResponse, model.sourcePath, model.destPath);
	} else {
		return A3(author$project$Commands$doCopy, author$project$Modals$MoveCopy$GotActionResponse, model.sourcePath, model.destPath);
	}
};
var elm$core$List$takeReverse = F3(
	function (n, list, kept) {
		takeReverse:
		while (true) {
			if (n <= 0) {
				return kept;
			} else {
				if (!list.b) {
					return kept;
				} else {
					var x = list.a;
					var xs = list.b;
					var $temp$n = n - 1,
						$temp$list = xs,
						$temp$kept = A2(elm$core$List$cons, x, kept);
					n = $temp$n;
					list = $temp$list;
					kept = $temp$kept;
					continue takeReverse;
				}
			}
		}
	});
var elm$core$List$takeTailRec = F2(
	function (n, list) {
		return elm$core$List$reverse(
			A3(elm$core$List$takeReverse, n, list, _List_Nil));
	});
var elm$core$List$takeFast = F3(
	function (ctr, n, list) {
		if (n <= 0) {
			return _List_Nil;
		} else {
			var _n0 = _Utils_Tuple2(n, list);
			_n0$1:
			while (true) {
				_n0$5:
				while (true) {
					if (!_n0.b.b) {
						return list;
					} else {
						if (_n0.b.b.b) {
							switch (_n0.a) {
								case 1:
									break _n0$1;
								case 2:
									var _n2 = _n0.b;
									var x = _n2.a;
									var _n3 = _n2.b;
									var y = _n3.a;
									return _List_fromArray(
										[x, y]);
								case 3:
									if (_n0.b.b.b.b) {
										var _n4 = _n0.b;
										var x = _n4.a;
										var _n5 = _n4.b;
										var y = _n5.a;
										var _n6 = _n5.b;
										var z = _n6.a;
										return _List_fromArray(
											[x, y, z]);
									} else {
										break _n0$5;
									}
								default:
									if (_n0.b.b.b.b && _n0.b.b.b.b.b) {
										var _n7 = _n0.b;
										var x = _n7.a;
										var _n8 = _n7.b;
										var y = _n8.a;
										var _n9 = _n8.b;
										var z = _n9.a;
										var _n10 = _n9.b;
										var w = _n10.a;
										var tl = _n10.b;
										return (ctr > 1000) ? A2(
											elm$core$List$cons,
											x,
											A2(
												elm$core$List$cons,
												y,
												A2(
													elm$core$List$cons,
													z,
													A2(
														elm$core$List$cons,
														w,
														A2(elm$core$List$takeTailRec, n - 4, tl))))) : A2(
											elm$core$List$cons,
											x,
											A2(
												elm$core$List$cons,
												y,
												A2(
													elm$core$List$cons,
													z,
													A2(
														elm$core$List$cons,
														w,
														A3(elm$core$List$takeFast, ctr + 1, n - 4, tl)))));
									} else {
										break _n0$5;
									}
							}
						} else {
							if (_n0.a === 1) {
								break _n0$1;
							} else {
								break _n0$5;
							}
						}
					}
				}
				return list;
			}
			var _n1 = _n0.b;
			var x = _n1.a;
			return _List_fromArray(
				[x]);
		}
	});
var elm$core$List$take = F2(
	function (n, list) {
		return A3(elm$core$List$takeFast, 0, n, list);
	});
var author$project$Util$dirname = function (path) {
	var split = author$project$Util$splitPath(path);
	if (!split.b) {
		return '/';
	} else {
		return author$project$Util$joinPath(
			A2(
				elm$core$List$take,
				elm$core$List$length(split) - 1,
				split));
	}
};
var elm$core$Basics$neq = _Utils_notEqual;
var elm$core$Basics$not = _Basics_not;
var author$project$Modals$MoveCopy$filterInvalidTargets = F2(
	function (sourcePath, path) {
		return (!_Utils_eq(
			path,
			author$project$Util$dirname(sourcePath))) && (!A2(elm$core$String$startsWith, path, sourcePath));
	});
var author$project$Modals$MoveCopy$fixPath = function (path) {
	return (path === '/') ? 'Home' : A2(
		elm$core$String$join,
		'/',
		author$project$Util$splitPath(path));
};
var author$project$Modals$MoveCopy$fixAllDirResponse = F2(
	function (model, paths) {
		return A2(
			elm$core$List$map,
			author$project$Modals$MoveCopy$fixPath,
			A2(
				elm$core$List$filter,
				author$project$Modals$MoveCopy$filterInvalidTargets(model.sourcePath),
				paths));
	});
var author$project$Modals$MoveCopy$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'DoAction':
				return _Utils_Tuple2(
					model,
					author$project$Modals$MoveCopy$doAction(model));
			case 'DirChosen':
				var path = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{destPath: path}),
					elm$core$Platform$Cmd$none);
			case 'SearchInput':
				var filter = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{filter: filter}),
					elm$core$Platform$Cmd$none);
			case 'GotAllDirsResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var dirs = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$MoveCopy$Ready(
									A2(author$project$Modals$MoveCopy$fixAllDirResponse, model, dirs))
							}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$MoveCopy$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotActionResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$MoveCopy$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				var sourcePath = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{destPath: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$shown, sourcePath: sourcePath, state: author$project$Modals$MoveCopy$Loading}),
					author$project$Commands$doListAllDirs(author$project$Modals$MoveCopy$GotAllDirsResponse));
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			default:
				var key = msg.a;
				return _Utils_Tuple2(
					model,
					function () {
						if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden) || (model.destPath === '')) {
							return elm$core$Platform$Cmd$none;
						} else {
							if (key === 'Enter') {
								return author$project$Modals$MoveCopy$doAction(model);
							} else {
								return elm$core$Platform$Cmd$none;
							}
						}
					}());
		}
	});
var author$project$Modals$Remove$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Modals$Remove$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var author$project$Modals$Remove$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'RemoveAll':
				var paths = msg.a;
				return _Utils_Tuple2(
					model,
					A2(author$project$Commands$doRemove, author$project$Modals$Remove$GotResponse, paths));
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Remove$Ready}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$Remove$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				var paths = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$shown, selected: paths}),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Remove$Ready}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			default:
				var key = msg.a;
				return _Utils_Tuple2(
					model,
					function () {
						if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden)) {
							return elm$core$Platform$Cmd$none;
						} else {
							if (key === 'Enter') {
								return A2(author$project$Commands$doRemove, author$project$Modals$Remove$GotResponse, model.selected);
							} else {
								return elm$core$Platform$Cmd$none;
							}
						}
					}());
		}
	});
var author$project$Modals$Rename$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Modals$Rename$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var author$project$Util$basename = function (path) {
	var split = elm$core$List$reverse(
		author$project$Util$splitPath(path));
	if (!split.b) {
		return '/';
	} else {
		var x = split.a;
		return x;
	}
};
var author$project$Modals$Rename$triggerRename = F2(
	function (sourcePath, newName) {
		return A3(
			author$project$Commands$doMove,
			author$project$Modals$Rename$GotResponse,
			sourcePath,
			author$project$Util$joinPath(
				_List_fromArray(
					[
						author$project$Util$dirname(sourcePath),
						author$project$Util$basename(newName)
					])));
	});
var author$project$Modals$Rename$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'DoRename':
				return _Utils_Tuple2(
					model,
					A2(author$project$Modals$Rename$triggerRename, model.currPath, model.inputName));
			case 'InputChanged':
				var inputName = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{inputName: inputName}),
					elm$core$Platform$Cmd$none);
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Rename$Ready}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$Rename$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				var currPath = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{currPath: currPath, inputName: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$shown}),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$Rename$Ready}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			default:
				var key = msg.a;
				if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden)) {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					if (key === 'Enter') {
						return _Utils_Tuple2(
							model,
							A2(author$project$Modals$Rename$triggerRename, model.currPath, model.inputName));
					} else {
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
					}
				}
		}
	});
var author$project$Modals$Share$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				var paths = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$shown, paths: paths}),
					elm$core$Platform$Cmd$none);
			default:
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, paths: _List_Nil}),
					elm$core$Platform$Cmd$none);
		}
	});
var elm$file$File$name = _File_name;
var elm$http$Http$expectBytesResponse = F2(
	function (toMsg, toResult) {
		return A3(
			_Http_expect,
			'arraybuffer',
			_Http_toDataView,
			A2(elm$core$Basics$composeR, toResult, toMsg));
	});
var elm$http$Http$expectWhatever = function (toMsg) {
	return A2(
		elm$http$Http$expectBytesResponse,
		toMsg,
		elm$http$Http$resolve(
			function (_n0) {
				return elm$core$Result$Ok(_Utils_Tuple0);
			}));
};
var elm$http$Http$filePart = _Http_pair;
var elm$http$Http$multipartBody = function (parts) {
	return A2(
		_Http_pair,
		'',
		_Http_toFormData(parts));
};
var elm$url$Url$percentEncode = _Url_percentEncode;
var author$project$Commands$doUpload = F3(
	function (toMsg, destPath, file) {
		return elm$http$Http$request(
			{
				body: elm$http$Http$multipartBody(
					_List_fromArray(
						[
							A2(elm$http$Http$filePart, 'files[]', file)
						])),
				expect: elm$http$Http$expectWhatever(
					toMsg(
						elm$file$File$name(file))),
				headers: _List_Nil,
				method: 'POST',
				timeout: elm$core$Maybe$Nothing,
				tracker: elm$core$Maybe$Just(
					'upload-' + elm$file$File$name(file)),
				url: '/api/v0/upload?root=' + elm$url$Url$percentEncode(destPath)
			});
	});
var author$project$Modals$Upload$Alertable = F2(
	function (alert, path) {
		return {alert: alert, path: path};
	});
var author$project$Modals$Upload$Uploaded = F2(
	function (a, b) {
		return {$: 'Uploaded', a: a, b: b};
	});
var author$project$Modals$Upload$alertMapper = F3(
	function (path, vis, a) {
		var _n0 = _Utils_eq(a.path, path);
		if (_n0) {
			return _Utils_update(
				a,
				{alert: vis});
		} else {
			return a;
		}
	});
var elm$http$Http$cancel = function (tracker) {
	return elm$http$Http$command(
		elm$http$Http$Cancel(tracker));
};
var elm$core$Basics$clamp = F3(
	function (low, high, number) {
		return (_Utils_cmp(number, low) < 0) ? low : ((_Utils_cmp(number, high) > 0) ? high : number);
	});
var elm$http$Http$fractionSent = function (p) {
	return (!p.size) ? 1 : A3(elm$core$Basics$clamp, 0, 1, p.sent / p.size);
};
var author$project$Modals$Upload$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'UploadSelectedFiles':
				var root = msg.a;
				var files = msg.b;
				var newUploads = A2(
					elm$core$Dict$union,
					model.uploads,
					elm$core$Dict$fromList(
						A2(
							elm$core$List$map,
							function (f) {
								return _Utils_Tuple2(
									elm$file$File$name(f),
									0);
							},
							files)));
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{uploads: newUploads}),
					elm$core$Platform$Cmd$batch(
						A2(
							elm$core$List$map,
							A2(author$project$Commands$doUpload, author$project$Modals$Upload$Uploaded, root),
							files)));
			case 'UploadProgress':
				var path = msg.a;
				var progress = msg.b;
				if (progress.$ === 'Sending') {
					var p = progress.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								uploads: A3(
									elm$core$Dict$insert,
									path,
									elm$http$Http$fractionSent(p),
									model.uploads)
							}),
						elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'Uploaded':
				var path = msg.a;
				var result = msg.b;
				var newUploads = A2(elm$core$Dict$remove, path, model.uploads);
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								success: A2(
									elm$core$List$cons,
									A2(author$project$Modals$Upload$Alertable, rundis$elm_bootstrap$Bootstrap$Alert$shown, path),
									model.success),
								uploads: newUploads
							}),
						A3(
							andrewMacmurray$elm_delay$Delay$after,
							5,
							andrewMacmurray$elm_delay$Delay$Second,
							A2(author$project$Modals$Upload$AlertMsg, path, rundis$elm_bootstrap$Bootstrap$Alert$closed)));
				} else {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								failed: A2(
									elm$core$List$cons,
									A2(author$project$Modals$Upload$Alertable, rundis$elm_bootstrap$Bootstrap$Alert$shown, path),
									model.failed),
								uploads: newUploads
							}),
						A3(
							andrewMacmurray$elm_delay$Delay$after,
							30,
							andrewMacmurray$elm_delay$Delay$Second,
							A2(author$project$Modals$Upload$AlertMsg, path, rundis$elm_bootstrap$Bootstrap$Alert$closed)));
				}
			case 'UploadCancel':
				var path = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							uploads: A2(elm$core$Dict$remove, path, model.uploads)
						}),
					elm$http$Http$cancel('upload-' + path));
			default:
				var path = msg.a;
				var vis = msg.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							failed: A2(
								elm$core$List$map,
								A2(author$project$Modals$Upload$alertMapper, path, vis),
								model.failed),
							success: A2(
								elm$core$List$map,
								A2(author$project$Modals$Upload$alertMapper, path, vis),
								model.success)
						}),
					elm$core$Platform$Cmd$none);
		}
	});
var author$project$Routes$Ls$Ascending = {$: 'Ascending'};
var author$project$Routes$Ls$Failure = {$: 'Failure'};
var author$project$Routes$Ls$GotPinResponse = function (a) {
	return {$: 'GotPinResponse', a: a};
};
var author$project$Routes$Ls$None = {$: 'None'};
var author$project$Routes$Ls$RemoveResponse = function (a) {
	return {$: 'RemoveResponse', a: a};
};
var author$project$Routes$Ls$Success = function (a) {
	return {$: 'Success', a: a};
};
var author$project$Routes$Ls$fixDropdownState = F3(
	function (refEntry, state, entry) {
		return _Utils_eq(entry.path, refEntry.path) ? _Utils_update(
			entry,
			{dropdown: state}) : entry;
	});
var author$project$Routes$Ls$setDropdownState = F3(
	function (model, entry, state) {
		var _n0 = model.state;
		if (_n0.$ === 'Success') {
			var actualModel = _n0.a;
			return _Utils_update(
				model,
				{
					state: author$project$Routes$Ls$Success(
						_Utils_update(
							actualModel,
							{
								entries: A2(
									elm$core$List$map,
									A2(author$project$Routes$Ls$fixDropdownState, entry, state),
									actualModel.entries)
							}))
				});
		} else {
			return model;
		}
	});
var author$project$Routes$Ls$Descending = {$: 'Descending'};
var author$project$Routes$Ls$entryPinToSortKey = function (entry) {
	var _n0 = _Utils_Tuple2(entry.isPinned, entry.isExplicit);
	if (_n0.a) {
		if (_n0.b) {
			return 2;
		} else {
			return 1;
		}
	} else {
		return 0;
	}
};
var elm$core$List$sortBy = _List_sortBy;
var elm$core$String$toLower = _String_toLower;
var elm$time$Time$posixToMillis = function (_n0) {
	var millis = _n0.a;
	return millis;
};
var author$project$Routes$Ls$sortByAscending = F2(
	function (model, key) {
		switch (key.$) {
			case 'Name':
				return A2(
					elm$core$List$sortBy,
					function (e) {
						return elm$core$String$toLower(
							author$project$Util$basename(e.path));
					},
					model.entries);
			case 'ModTime':
				return A2(
					elm$core$List$sortBy,
					function (e) {
						return elm$time$Time$posixToMillis(e.lastModified);
					},
					model.entries);
			case 'Pin':
				return A2(
					elm$core$List$sortBy,
					function (e) {
						return author$project$Routes$Ls$entryPinToSortKey(e);
					},
					model.entries);
			case 'Size':
				return A2(
					elm$core$List$sortBy,
					function ($) {
						return $.size;
					},
					model.entries);
			default:
				return model.entries;
		}
	});
var author$project$Routes$Ls$sortBy = F3(
	function (model, direction, key) {
		if (direction.$ === 'Ascending') {
			return _Utils_update(
				model,
				{
					entries: A2(author$project$Routes$Ls$sortByAscending, model, key),
					sortState: _Utils_Tuple2(author$project$Routes$Ls$Ascending, key)
				});
		} else {
			return _Utils_update(
				model,
				{
					entries: elm$core$List$reverse(
						A2(author$project$Routes$Ls$sortByAscending, model, key)),
					sortState: _Utils_Tuple2(author$project$Routes$Ls$Descending, key)
				});
		}
	});
var elm$core$Set$Set_elm_builtin = function (a) {
	return {$: 'Set_elm_builtin', a: a};
};
var elm$core$Set$insert = F2(
	function (key, _n0) {
		var dict = _n0.a;
		return elm$core$Set$Set_elm_builtin(
			A3(elm$core$Dict$insert, key, _Utils_Tuple0, dict));
	});
var elm$core$Set$remove = F2(
	function (key, _n0) {
		var dict = _n0.a;
		return elm$core$Set$Set_elm_builtin(
			A2(elm$core$Dict$remove, key, dict));
	});
var elm$core$Dict$sizeHelp = F2(
	function (n, dict) {
		sizeHelp:
		while (true) {
			if (dict.$ === 'RBEmpty_elm_builtin') {
				return n;
			} else {
				var left = dict.d;
				var right = dict.e;
				var $temp$n = A2(elm$core$Dict$sizeHelp, n + 1, right),
					$temp$dict = left;
				n = $temp$n;
				dict = $temp$dict;
				continue sizeHelp;
			}
		}
	});
var elm$core$Dict$size = function (dict) {
	return A2(elm$core$Dict$sizeHelp, 0, dict);
};
var elm$core$Set$size = function (_n0) {
	var dict = _n0.a;
	return elm$core$Dict$size(dict);
};
var author$project$Routes$Ls$updateCheckboxTickActual = F3(
	function (path, isChecked, model) {
		if (isChecked) {
			var updatedSet = A2(elm$core$Set$insert, path, model.checked);
			return _Utils_update(
				model,
				{
					checked: _Utils_eq(
						elm$core$Set$size(updatedSet),
						elm$core$List$length(model.entries)) ? A2(elm$core$Set$insert, '', updatedSet) : updatedSet
				});
		} else {
			return _Utils_update(
				model,
				{
					checked: A2(
						elm$core$Set$remove,
						'',
						A2(elm$core$Set$remove, path, model.checked))
				});
		}
	});
var author$project$Routes$Ls$updateCheckboxTick = F3(
	function (path, isChecked, model) {
		var _n0 = model.state;
		if (_n0.$ === 'Success') {
			var actualModel = _n0.a;
			return _Utils_update(
				model,
				{
					state: author$project$Routes$Ls$Success(
						A3(author$project$Routes$Ls$updateCheckboxTickActual, path, isChecked, actualModel))
				});
		} else {
			return model;
		}
	});
var elm$core$Set$empty = elm$core$Set$Set_elm_builtin(elm$core$Dict$empty);
var elm$core$Set$fromList = function (list) {
	return A3(elm$core$List$foldl, elm$core$Set$insert, elm$core$Set$empty, list);
};
var author$project$Routes$Ls$updateCheckboxTickAllActual = F2(
	function (isChecked, model) {
		if (isChecked) {
			return _Utils_update(
				model,
				{
					checked: elm$core$Set$fromList(
						_Utils_ap(
							A2(
								elm$core$List$map,
								function (e) {
									return e.path;
								},
								model.entries),
							_List_fromArray(
								[''])))
				});
		} else {
			return _Utils_update(
				model,
				{checked: elm$core$Set$empty});
		}
	});
var author$project$Routes$Ls$updateCheckboxTickAll = F2(
	function (isChecked, model) {
		var _n0 = model.state;
		if (_n0.$ === 'Success') {
			var actualModel = _n0.a;
			return _Utils_update(
				model,
				{
					state: author$project$Routes$Ls$Success(
						A2(author$project$Routes$Ls$updateCheckboxTickAllActual, isChecked, actualModel))
				});
		} else {
			return model;
		}
	});
var author$project$Util$urlEncodePath = function (path) {
	return author$project$Util$joinPath(
		A2(
			elm$core$List$map,
			elm$url$Url$percentEncode,
			author$project$Util$splitPath(path)));
};
var elm$browser$Browser$Navigation$pushUrl = _Browser_pushUrl;
var elm$core$Dict$singleton = F2(
	function (key, value) {
		return A5(elm$core$Dict$RBNode_elm_builtin, elm$core$Dict$Black, key, value, elm$core$Dict$RBEmpty_elm_builtin, elm$core$Dict$RBEmpty_elm_builtin);
	});
var elm$core$Set$singleton = function (key) {
	return elm$core$Set$Set_elm_builtin(
		A2(elm$core$Dict$singleton, key, _Utils_Tuple0));
};
var elm$url$Url$Builder$QueryParameter = F2(
	function (a, b) {
		return {$: 'QueryParameter', a: a, b: b};
	});
var elm$url$Url$Builder$string = F2(
	function (key, value) {
		return A2(
			elm$url$Url$Builder$QueryParameter,
			elm$url$Url$percentEncode(key),
			elm$url$Url$percentEncode(value));
	});
var elm$url$Url$Builder$toQueryPair = function (_n0) {
	var key = _n0.a;
	var value = _n0.b;
	return key + ('=' + value);
};
var elm$url$Url$Builder$toQuery = function (parameters) {
	if (!parameters.b) {
		return '';
	} else {
		return '?' + A2(
			elm$core$String$join,
			'&',
			A2(elm$core$List$map, elm$url$Url$Builder$toQueryPair, parameters));
	}
};
var author$project$Routes$Ls$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'ActionDropdownMsg':
				var entry = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					A3(author$project$Routes$Ls$setDropdownState, model, entry, state),
					elm$core$Platform$Cmd$none);
			case 'RowClicked':
				var entry = msg.a;
				return _Utils_Tuple2(
					model,
					A2(
						elm$browser$Browser$Navigation$pushUrl,
						model.key,
						'/view' + author$project$Util$urlEncodePath(entry.path)));
			case 'RemoveClicked':
				var entry = msg.a;
				return _Utils_Tuple2(
					A3(author$project$Routes$Ls$setDropdownState, model, entry, rundis$elm_bootstrap$Bootstrap$Dropdown$initialState),
					A2(
						author$project$Commands$doRemove,
						author$project$Routes$Ls$RemoveResponse,
						_List_fromArray(
							[entry.path])));
			case 'SearchInput':
				var query = msg.a;
				return _Utils_Tuple2(
					model,
					A2(
						elm$browser$Browser$Navigation$pushUrl,
						model.key,
						_Utils_ap(
							model.url.path,
							(!elm$core$String$length(query)) ? '' : elm$url$Url$Builder$toQuery(
								_List_fromArray(
									[
										A2(elm$url$Url$Builder$string, 'filter', query)
									])))));
			case 'HistoryClicked':
				var entry = msg.a;
				return _Utils_Tuple2(
					A3(author$project$Routes$Ls$setDropdownState, model, entry, rundis$elm_bootstrap$Bootstrap$Dropdown$initialState),
					A2(
						elm$core$Platform$Cmd$map,
						author$project$Routes$Ls$HistoryMsg,
						author$project$Modals$History$show(entry.path)));
			case 'SortBy':
				var direction = msg.a;
				var key = msg.b;
				var _n1 = model.state;
				if (_n1.$ === 'Success') {
					var actualModel = _n1.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Routes$Ls$Success(
									A3(author$project$Routes$Ls$sortBy, actualModel, direction, key))
							}),
						elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'RemoveResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								alert: rundis$elm_bootstrap$Bootstrap$Alert$shown,
								currError: author$project$Util$httpErrorToString(err)
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var response = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Routes$Ls$Success(
									{
										checked: response.self.isDir ? elm$core$Set$empty : elm$core$Set$singleton(response.self.path),
										entries: response.entries,
										isFiltered: response.isFiltered,
										self: response.self,
										sortState: _Utils_Tuple2(author$project$Routes$Ls$Ascending, author$project$Routes$Ls$None)
									})
							}),
						elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{state: author$project$Routes$Ls$Failure}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotPinResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'CheckboxTick':
				var path = msg.a;
				var isChecked = msg.b;
				return _Utils_Tuple2(
					A3(author$project$Routes$Ls$updateCheckboxTick, path, isChecked, model),
					elm$core$Platform$Cmd$none);
			case 'CheckboxTickAll':
				var isChecked = msg.a;
				return _Utils_Tuple2(
					A2(author$project$Routes$Ls$updateCheckboxTickAll, isChecked, model),
					elm$core$Platform$Cmd$none);
			case 'PinClicked':
				var path = msg.a;
				var shouldBePinned = msg.b;
				return shouldBePinned ? _Utils_Tuple2(
					model,
					A3(author$project$Commands$doPin, author$project$Routes$Ls$GotPinResponse, path, 'curr')) : _Utils_Tuple2(
					model,
					A3(author$project$Commands$doUnpin, author$project$Routes$Ls$GotPinResponse, path, 'curr'));
			case 'AlertMsg':
				var state = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: state}),
					elm$core$Platform$Cmd$none);
			case 'HistoryMsg':
				var subMsg = msg.a;
				var _n5 = A2(author$project$Modals$History$update, subMsg, model.historyState);
				var newSubModel = _n5.a;
				var newSubCmd = _n5.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{historyState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$HistoryMsg, newSubCmd));
			case 'RenameMsg':
				var subMsg = msg.a;
				var _n6 = A2(author$project$Modals$Rename$update, subMsg, model.renameState);
				var newSubModel = _n6.a;
				var newSubCmd = _n6.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{renameState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$RenameMsg, newSubCmd));
			case 'MoveMsg':
				var subMsg = msg.a;
				var _n7 = A2(author$project$Modals$MoveCopy$update, subMsg, model.moveState);
				var newSubModel = _n7.a;
				var newSubCmd = _n7.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{moveState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$MoveMsg, newSubCmd));
			case 'CopyMsg':
				var subMsg = msg.a;
				var _n8 = A2(author$project$Modals$MoveCopy$update, subMsg, model.copyState);
				var newSubModel = _n8.a;
				var newSubCmd = _n8.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{copyState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$CopyMsg, newSubCmd));
			case 'UploadMsg':
				var subMsg = msg.a;
				var _n9 = A2(author$project$Modals$Upload$update, subMsg, model.uploadState);
				var newSubModel = _n9.a;
				var newSubCmd = _n9.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{uploadState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$UploadMsg, newSubCmd));
			case 'MkdirMsg':
				var subMsg = msg.a;
				var _n10 = A2(author$project$Modals$Mkdir$update, subMsg, model.mkdirState);
				var newSubModel = _n10.a;
				var newSubCmd = _n10.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{mkdirState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$MkdirMsg, newSubCmd));
			case 'RemoveMsg':
				var subMsg = msg.a;
				var _n11 = A2(author$project$Modals$Remove$update, subMsg, model.removeState);
				var newSubModel = _n11.a;
				var newSubCmd = _n11.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{removeState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$RemoveMsg, newSubCmd));
			default:
				var subMsg = msg.a;
				var _n12 = A2(author$project$Modals$Share$update, subMsg, model.shareState);
				var newSubModel = _n12.a;
				var newSubCmd = _n12.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{shareState: newSubModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Ls$ShareMsg, newSubCmd));
		}
	});
var author$project$Commands$decodeRemoteAddQuery = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var elm$json$Json$Encode$bool = _Json_wrap;
var author$project$Commands$encodeFolder = function (f) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'folder',
				elm$json$Json$Encode$string(f.folder)),
				_Utils_Tuple2(
				'read_only',
				elm$json$Json$Encode$bool(f.readOnly)),
				_Utils_Tuple2(
				'conflict_strategy',
				elm$json$Json$Encode$string(f.conflictStrategy))
			]));
};
var author$project$Commands$encodeRemoteAddQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'name',
				elm$json$Json$Encode$string(q.name)),
				_Utils_Tuple2(
				'fingerprint',
				elm$json$Json$Encode$string(q.fingerprint)),
				_Utils_Tuple2(
				'accept_auto_updates',
				elm$json$Json$Encode$bool(q.doAutoUpdate)),
				_Utils_Tuple2(
				'folders',
				A2(elm$json$Json$Encode$list, author$project$Commands$encodeFolder, q.folders)),
				_Utils_Tuple2(
				'accept_push',
				elm$json$Json$Encode$bool(q.acceptPush)),
				_Utils_Tuple2(
				'conflict_strategy',
				elm$json$Json$Encode$string(q.conflictStrategy))
			]));
};
var author$project$Commands$doRemoteModify = F2(
	function (toMsg, remote) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeRemoteAddQuery(
						{acceptPush: remote.acceptPush, conflictStrategy: remote.conflictStrategy, doAutoUpdate: remote.acceptAutoUpdates, fingerprint: remote.fingerprint, folders: remote.folders, name: remote.name})),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeRemoteAddQuery),
				url: '/api/v0/remotes/modify'
			});
	});
var author$project$Commands$RemoteSyncQuery = function (name) {
	return {name: name};
};
var author$project$Commands$decodeRemoteSyncQuery = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodeRemoteSyncQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'name',
				elm$json$Json$Encode$string(q.name))
			]));
};
var author$project$Commands$doRemoteSync = F2(
	function (toMsg, name) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeRemoteSyncQuery(
						author$project$Commands$RemoteSyncQuery(name))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeRemoteSyncQuery),
				url: '/api/v0/remotes/sync'
			});
	});
var author$project$Modals$RemoteAdd$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Commands$doRemoteAdd = F7(
	function (toMsg, name, fingerprint, doAutoUpdate, acceptPush, conflictStrategy, folders) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeRemoteAddQuery(
						{acceptPush: acceptPush, conflictStrategy: conflictStrategy, doAutoUpdate: doAutoUpdate, fingerprint: fingerprint, folders: folders, name: name})),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeRemoteAddQuery),
				url: '/api/v0/remotes/add'
			});
	});
var author$project$Modals$RemoteAdd$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var author$project$Modals$RemoteAdd$submit = function (model) {
	return A7(author$project$Commands$doRemoteAdd, author$project$Modals$RemoteAdd$GotResponse, model.name, model.fingerprint, model.doAutoUdate, model.acceptPush, model.conflictStrategy, _List_Nil);
};
var author$project$Modals$RemoteAdd$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'RemoteAdd':
				return _Utils_Tuple2(
					model,
					author$project$Modals$RemoteAdd$submit(model));
			case 'NameInputChanged':
				var name = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{name: name}),
					elm$core$Platform$Cmd$none);
			case 'FingerprintInputChanged':
				var fingerprint = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{fingerprint: fingerprint}),
					elm$core$Platform$Cmd$none);
			case 'AutoUpdateChanged':
				var doAutoUdate = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{doAutoUdate: doAutoUdate}),
					elm$core$Platform$Cmd$none);
			case 'AcceptPushChanged':
				var acceptPush = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{acceptPush: acceptPush}),
					elm$core$Platform$Cmd$none);
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden, state: author$project$Modals$RemoteAdd$Ready}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$RemoteAdd$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				return _Utils_Tuple2(
					author$project$Modals$RemoteAdd$newModelWithState(rundis$elm_bootstrap$Bootstrap$Modal$shown),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			case 'ConflictDropdownMsg':
				var state = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{conflictDropdown: state}),
					elm$core$Platform$Cmd$none);
			case 'ConflictStrategyChanged':
				var state = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{conflictStrategy: state}),
					elm$core$Platform$Cmd$none);
			default:
				var key = msg.a;
				if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden)) {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					if (key === 'Enter') {
						return _Utils_Tuple2(
							model,
							author$project$Modals$RemoteAdd$submit(model));
					} else {
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
					}
				}
		}
	});
var author$project$Modals$RemoteFolders$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Modals$RemoteFolders$GotAllDirsResponse = function (a) {
	return {$: 'GotAllDirsResponse', a: a};
};
var author$project$Modals$RemoteFolders$fixFolder = function (path) {
	return author$project$Util$prefixSlash(path);
};
var author$project$Modals$RemoteFolders$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var author$project$Modals$RemoteFolders$submit = function (remote) {
	return A2(author$project$Commands$doRemoteModify, author$project$Modals$RemoteFolders$GotResponse, remote);
};
var author$project$Modals$RemoteFolders$addFolder = F2(
	function (model, folder) {
		var oldRemote = model.remote;
		var cleanFolder = A3(
			author$project$Commands$Folder,
			author$project$Modals$RemoteFolders$fixFolder(folder),
			false,
			'');
		var newRemote = _Utils_update(
			oldRemote,
			{
				folders: A2(
					elm$core$List$sortBy,
					function ($) {
						return $.folder;
					},
					A2(elm$core$List$cons, cleanFolder, oldRemote.folders))
			});
		var upModel = _Utils_update(
			model,
			{remote: newRemote});
		return _Utils_Tuple2(
			upModel,
			author$project$Modals$RemoteFolders$submit(upModel.remote));
	});
var author$project$Modals$RemoteFolders$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{state: author$project$Modals$RemoteFolders$Ready}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$RemoteFolders$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'FolderRemove':
				var folder = msg.a;
				var oldRemote = model.remote;
				var newRemote = _Utils_update(
					oldRemote,
					{
						folders: A2(
							elm$core$List$filter,
							function (f) {
								return !_Utils_eq(f.folder, folder);
							},
							oldRemote.folders)
					});
				var upModel = _Utils_update(
					model,
					{remote: newRemote});
				return _Utils_Tuple2(
					upModel,
					author$project$Modals$RemoteFolders$submit(upModel.remote));
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				var remote = msg.a;
				return _Utils_Tuple2(
					A2(author$project$Modals$RemoteFolders$newModelWithState, rundis$elm_bootstrap$Bootstrap$Modal$shown, remote),
					author$project$Commands$doListAllDirs(author$project$Modals$RemoteFolders$GotAllDirsResponse));
			case 'GotAllDirsResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var allDirs = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{allDirs: allDirs}),
						elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'DirChosen':
				var choice = msg.a;
				return A2(author$project$Modals$RemoteFolders$addFolder, model, choice);
			case 'SearchInput':
				var filter = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{filter: filter}),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{filter: '', modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			case 'ReadOnlyChanged':
				var path = msg.a;
				var state = msg.b;
				var oldRemote = model.remote;
				var newRemote = _Utils_update(
					oldRemote,
					{
						folders: A2(
							elm$core$List$map,
							function (f) {
								return _Utils_eq(f.folder, path) ? _Utils_update(
									f,
									{readOnly: state}) : f;
							},
							model.remote.folders)
					});
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{remote: newRemote}),
					author$project$Modals$RemoteFolders$submit(newRemote));
			case 'ConflictDropdownMsg':
				var folder = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							conflictDropdowns: A3(elm$core$Dict$insert, folder, state, model.conflictDropdowns)
						}),
					elm$core$Platform$Cmd$none);
			default:
				var folder = msg.a;
				var strategy = msg.b;
				var oldRemote = model.remote;
				var newFolders = A2(
					elm$core$List$map,
					function (f) {
						return _Utils_eq(f.folder, folder) ? _Utils_update(
							f,
							{conflictStrategy: strategy}) : f;
					},
					model.remote.folders);
				var newRemote = _Utils_update(
					oldRemote,
					{folders: newFolders});
				var upModel = _Utils_update(
					model,
					{remote: newRemote});
				return _Utils_Tuple2(
					upModel,
					author$project$Modals$RemoteFolders$submit(upModel.remote));
		}
	});
var author$project$Modals$RemoteRemove$Fail = function (a) {
	return {$: 'Fail', a: a};
};
var author$project$Commands$RemoteRemoveQuery = function (name) {
	return {name: name};
};
var author$project$Commands$decodeRemoteRemoveQuery = A2(elm$json$Json$Decode$field, 'message', elm$json$Json$Decode$string);
var author$project$Commands$encodeRemoteRemoveQuery = function (q) {
	return elm$json$Json$Encode$object(
		_List_fromArray(
			[
				_Utils_Tuple2(
				'name',
				elm$json$Json$Encode$string(q.name))
			]));
};
var author$project$Commands$doRemoteRemove = F2(
	function (toMsg, name) {
		return elm$http$Http$post(
			{
				body: elm$http$Http$jsonBody(
					author$project$Commands$encodeRemoteRemoveQuery(
						author$project$Commands$RemoteRemoveQuery(name))),
				expect: A2(elm$http$Http$expectJson, toMsg, author$project$Commands$decodeRemoteRemoveQuery),
				url: '/api/v0/remotes/remove'
			});
	});
var author$project$Modals$RemoteRemove$GotResponse = function (a) {
	return {$: 'GotResponse', a: a};
};
var author$project$Modals$RemoteRemove$submit = function (model) {
	return A2(author$project$Commands$doRemoteRemove, author$project$Modals$RemoteRemove$GotResponse, model.name);
};
var author$project$Modals$RemoteRemove$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'DoRemove':
				return _Utils_Tuple2(
					model,
					author$project$Modals$RemoteRemove$submit(model));
			case 'GotResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Modals$RemoteRemove$Fail(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'AnimateModal':
				var visibility = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: visibility}),
					elm$core$Platform$Cmd$none);
			case 'ModalShow':
				var path = msg.a;
				return _Utils_Tuple2(
					A2(author$project$Modals$RemoteRemove$newModelWithState, path, rundis$elm_bootstrap$Bootstrap$Modal$shown),
					elm$core$Platform$Cmd$none);
			case 'ModalClose':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{modal: rundis$elm_bootstrap$Bootstrap$Modal$hidden}),
					elm$core$Platform$Cmd$none);
			case 'AlertMsg':
				var vis = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: vis}),
					elm$core$Platform$Cmd$none);
			default:
				var key = msg.a;
				if (_Utils_eq(model.modal, rundis$elm_bootstrap$Bootstrap$Modal$hidden)) {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					if (key === 'Enter') {
						return _Utils_Tuple2(
							model,
							author$project$Modals$RemoteRemove$submit(model));
					} else {
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
					}
				}
		}
	});
var author$project$Routes$Remotes$Failure = function (a) {
	return {$: 'Failure', a: a};
};
var author$project$Routes$Remotes$GotRemoteModifyResponse = function (a) {
	return {$: 'GotRemoteModifyResponse', a: a};
};
var author$project$Routes$Remotes$GotSyncResponse = function (a) {
	return {$: 'GotSyncResponse', a: a};
};
var author$project$Routes$Remotes$Success = function (a) {
	return {$: 'Success', a: a};
};
var author$project$Routes$Remotes$showAlert = F4(
	function (model, duration, modalTyp, message) {
		var newAlert = A3(author$project$Util$AlertState, message, modalTyp, rundis$elm_bootstrap$Bootstrap$Alert$shown);
		return _Utils_Tuple2(
			_Utils_update(
				model,
				{alert: newAlert}),
			elm$core$Platform$Cmd$batch(
				_List_fromArray(
					[
						A3(
						andrewMacmurray$elm_delay$Delay$after,
						duration,
						andrewMacmurray$elm_delay$Delay$Second,
						author$project$Routes$Remotes$AlertMsg(rundis$elm_bootstrap$Bootstrap$Alert$closed))
					])));
	});
var author$project$Routes$Remotes$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'GotRemoteListResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var remotes = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Routes$Remotes$Success(remotes)
							}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								state: author$project$Routes$Remotes$Failure(
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotSyncResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return A4(author$project$Routes$Remotes$showAlert, model, 5, author$project$Util$Success, 'Succesfully synchronized!');
				} else {
					var err = result.a;
					return A4(
						author$project$Routes$Remotes$showAlert,
						model,
						20,
						author$project$Util$Danger,
						'Failed to sync: ' + author$project$Util$httpErrorToString(err));
				}
			case 'GotRemoteModifyResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return A4(
						author$project$Routes$Remotes$showAlert,
						model,
						20,
						author$project$Util$Danger,
						'Failed to set auto update: ' + author$project$Util$httpErrorToString(err));
				}
			case 'GotSelfResponse':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var self = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{self: self}),
						elm$core$Platform$Cmd$none);
				} else {
					var err = result.a;
					return A4(
						author$project$Routes$Remotes$showAlert,
						model,
						20,
						author$project$Util$Danger,
						'Failed to get information about ourselves: ' + author$project$Util$httpErrorToString(err));
				}
			case 'ActionDropdownMsg':
				var name = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							actionDropdowns: A3(elm$core$Dict$insert, name, state, model.actionDropdowns)
						}),
					elm$core$Platform$Cmd$none);
			case 'ConflictDropdownMsg':
				var name = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							conflictDropdowns: A3(elm$core$Dict$insert, name, state, model.conflictDropdowns)
						}),
					elm$core$Platform$Cmd$none);
			case 'SyncClicked':
				var name = msg.a;
				return _Utils_Tuple2(
					model,
					A2(author$project$Commands$doRemoteSync, author$project$Routes$Remotes$GotSyncResponse, name));
			case 'AutoUpdateToggled':
				var remote = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					model,
					A2(
						author$project$Commands$doRemoteModify,
						author$project$Routes$Remotes$GotRemoteModifyResponse,
						_Utils_update(
							remote,
							{acceptAutoUpdates: state})));
			case 'AcceptPushToggled':
				var remote = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					model,
					A2(
						author$project$Commands$doRemoteModify,
						author$project$Routes$Remotes$GotRemoteModifyResponse,
						_Utils_update(
							remote,
							{acceptPush: state})));
			case 'ConflictStrategyToggled':
				var remote = msg.a;
				var state = msg.b;
				return _Utils_Tuple2(
					model,
					A2(
						author$project$Commands$doRemoteModify,
						author$project$Routes$Remotes$GotRemoteModifyResponse,
						_Utils_update(
							remote,
							{conflictStrategy: state})));
			case 'RemoteAddMsg':
				var subMsg = msg.a;
				var _n5 = A2(author$project$Modals$RemoteAdd$update, subMsg, model.remoteAddState);
				var upModel = _n5.a;
				var upCmd = _n5.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{remoteAddState: upModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Remotes$RemoteAddMsg, upCmd));
			case 'RemoteRemoveMsg':
				var subMsg = msg.a;
				var _n6 = A2(author$project$Modals$RemoteRemove$update, subMsg, model.remoteRemoveState);
				var upModel = _n6.a;
				var upCmd = _n6.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{remoteRemoveState: upModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Remotes$RemoteRemoveMsg, upCmd));
			case 'RemoteFolderMsg':
				var subMsg = msg.a;
				var _n7 = A2(author$project$Modals$RemoteFolders$update, subMsg, model.remoteFoldersState);
				var upModel = _n7.a;
				var upCmd = _n7.b;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{remoteFoldersState: upModel}),
					A2(elm$core$Platform$Cmd$map, author$project$Routes$Remotes$RemoteFolderMsg, upCmd));
			default:
				var vis = msg.a;
				var newAlert = A3(author$project$Util$AlertState, model.alert.message, model.alert.typ, vis);
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{alert: newAlert}),
					elm$core$Platform$Cmd$none);
		}
	});
var elm$browser$Browser$Navigation$load = _Browser_load;
var elm$url$Url$addPort = F2(
	function (maybePort, starter) {
		if (maybePort.$ === 'Nothing') {
			return starter;
		} else {
			var port_ = maybePort.a;
			return starter + (':' + elm$core$String$fromInt(port_));
		}
	});
var elm$url$Url$addPrefixed = F3(
	function (prefix, maybeSegment, starter) {
		if (maybeSegment.$ === 'Nothing') {
			return starter;
		} else {
			var segment = maybeSegment.a;
			return _Utils_ap(
				starter,
				_Utils_ap(prefix, segment));
		}
	});
var elm$url$Url$toString = function (url) {
	var http = function () {
		var _n0 = url.protocol;
		if (_n0.$ === 'Http') {
			return 'http://';
		} else {
			return 'https://';
		}
	}();
	return A3(
		elm$url$Url$addPrefixed,
		'#',
		url.fragment,
		A3(
			elm$url$Url$addPrefixed,
			'?',
			url.query,
			_Utils_ap(
				A2(
					elm$url$Url$addPort,
					url.port_,
					_Utils_ap(http, url.host)),
				url.path)));
};
var author$project$Main$update = F2(
	function (msg, model) {
		switch (msg.$) {
			case 'AdjustTimeZone':
				var newZone = msg.a;
				var _n1 = model.loginState;
				if (_n1.$ === 'LoginSuccess') {
					var viewState = _n1.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								loginState: author$project$Main$LoginSuccess(
									_Utils_update(
										viewState,
										{
											listState: A2(author$project$Routes$Ls$changeTimeZone, newZone, viewState.listState)
										})),
								zone: newZone
							}),
						elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{zone: newZone}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotWhoamiResp':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var whoami = result.a;
					var _n3 = whoami.isLoggedIn;
					if (_n3) {
						return A5(author$project$Main$doInitAfterLogin, model, whoami.username, whoami.rights, whoami.isAnon, whoami.anonIsAllowed);
					} else {
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A2(author$project$Main$LoginReady, '', '')
								}),
							elm$core$Platform$Cmd$none);
					}
				} else {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								loginState: A2(author$project$Main$LoginReady, '', '')
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotLoginResp':
				var result = msg.a;
				if (result.$ === 'Ok') {
					var response = result.a;
					return A5(author$project$Main$doInitAfterLogin, model, response.username, response.rights, response.isAnon, response.anonIsAllowed);
				} else {
					var err = result.a;
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								loginState: A3(
									author$project$Main$LoginFailure,
									'',
									'',
									author$project$Util$httpErrorToString(err))
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'GotLogoutResp':
				var mayReloginAsAnon = msg.a;
				var _n5 = model.loginState;
				if (_n5.$ === 'LoginSuccess') {
					var viewState = _n5.a;
					return (mayReloginAsAnon && viewState.anonIsAllowed) ? _Utils_Tuple2(
						model,
						author$project$Commands$doWhoami(author$project$Main$GotWhoamiResp)) : _Utils_Tuple2(
						_Utils_update(
							model,
							{
								loginState: A2(author$project$Main$LoginReady, '', '')
							}),
						elm$core$Platform$Cmd$none);
				} else {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{
								loginState: A2(author$project$Main$LoginReady, '', '')
							}),
						elm$core$Platform$Cmd$none);
				}
			case 'LinkClicked':
				var urlRequest = msg.a;
				if (urlRequest.$ === 'Internal') {
					var url = urlRequest.a;
					var _n7 = A2(elm$core$String$startsWith, '/get', url.path);
					if (_n7) {
						return _Utils_Tuple2(
							model,
							elm$browser$Browser$Navigation$load(
								elm$url$Url$toString(url)));
					} else {
						return _Utils_Tuple2(
							model,
							A2(
								elm$browser$Browser$Navigation$pushUrl,
								model.key,
								elm$url$Url$toString(
									_Utils_update(
										url,
										{query: elm$core$Maybe$Nothing}))));
					}
				} else {
					var href = urlRequest.a;
					var currUrl = elm$url$Url$toString(model.url);
					switch (href) {
						case '':
							return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
						case '#':
							return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
						default:
							return _Utils_Tuple2(
								model,
								_Utils_eq(href, currUrl) ? elm$core$Platform$Cmd$none : elm$browser$Browser$Navigation$load(href));
					}
				}
			case 'UrlChanged':
				var url = msg.a;
				var _n9 = model.loginState;
				if (_n9.$ === 'LoginSuccess') {
					var viewState = _n9.a;
					var _n10 = A2(author$project$Main$viewFromUrl, viewState.rights, url);
					switch (_n10.$) {
						case 'ViewList':
							return _Utils_Tuple2(
								_Utils_update(
									model,
									{
										loginState: author$project$Main$LoginSuccess(
											_Utils_update(
												viewState,
												{
													commitsState: A2(author$project$Routes$Commits$updateUrl, viewState.commitsState, url),
													currentView: author$project$Main$ViewList,
													deletedFilesState: A2(author$project$Routes$DeletedFiles$updateUrl, viewState.deletedFilesState, url),
													diffState: A2(author$project$Routes$Diff$updateUrl, viewState.diffState, url),
													listState: A2(author$project$Routes$Ls$changeUrl, url, viewState.listState)
												})),
										url: url
									}),
								A2(
									elm$core$Platform$Cmd$map,
									author$project$Main$ListMsg,
									author$project$Routes$Ls$doListQueryFromUrl(url)));
						case 'ViewDiff':
							return _Utils_Tuple2(
								_Utils_update(
									model,
									{
										loginState: author$project$Main$LoginSuccess(
											_Utils_update(
												viewState,
												{
													commitsState: A2(author$project$Routes$Commits$updateUrl, viewState.commitsState, url),
													currentView: author$project$Main$ViewDiff,
													deletedFilesState: A2(author$project$Routes$DeletedFiles$updateUrl, viewState.deletedFilesState, url),
													diffState: A2(author$project$Routes$Diff$updateUrl, viewState.diffState, url)
												})),
										url: url
									}),
								A2(
									elm$core$Platform$Cmd$map,
									author$project$Main$DiffMsg,
									A2(author$project$Routes$Diff$reload, viewState.diffState, url)));
						case 'ViewCommits':
							return _Utils_Tuple2(
								_Utils_update(
									model,
									{
										loginState: author$project$Main$LoginSuccess(
											_Utils_update(
												viewState,
												{
													commitsState: A2(author$project$Routes$Commits$updateUrl, viewState.commitsState, url),
													currentView: author$project$Main$ViewCommits,
													deletedFilesState: A2(author$project$Routes$DeletedFiles$updateUrl, viewState.deletedFilesState, url),
													diffState: A2(author$project$Routes$Diff$updateUrl, viewState.diffState, url)
												})),
										url: url
									}),
								A2(
									elm$core$Platform$Cmd$map,
									author$project$Main$CommitsMsg,
									author$project$Routes$Commits$reloadIfNeeded(viewState.commitsState)));
						case 'ViewDeletedFiles':
							return _Utils_Tuple2(
								_Utils_update(
									model,
									{
										loginState: author$project$Main$LoginSuccess(
											_Utils_update(
												viewState,
												{
													commitsState: A2(author$project$Routes$Commits$updateUrl, viewState.commitsState, url),
													currentView: author$project$Main$ViewDeletedFiles,
													deletedFilesState: A2(author$project$Routes$DeletedFiles$updateUrl, viewState.deletedFilesState, url),
													diffState: A2(author$project$Routes$Diff$updateUrl, viewState.diffState, url)
												})),
										url: url
									}),
								A2(
									elm$core$Platform$Cmd$map,
									author$project$Main$DeletedFilesMsg,
									author$project$Routes$DeletedFiles$reloadIfNeeded(viewState.deletedFilesState)));
						default:
							var other = _n10;
							return _Utils_Tuple2(
								_Utils_update(
									model,
									{
										loginState: author$project$Main$LoginSuccess(
											_Utils_update(
												viewState,
												{
													commitsState: A2(author$project$Routes$Commits$updateUrl, viewState.commitsState, url),
													currentView: other,
													deletedFilesState: A2(author$project$Routes$DeletedFiles$updateUrl, viewState.deletedFilesState, url),
													diffState: A2(author$project$Routes$Diff$updateUrl, viewState.diffState, url)
												})),
										url: url
									}),
								elm$core$Platform$Cmd$none);
					}
				} else {
					return _Utils_Tuple2(
						_Utils_update(
							model,
							{url: url}),
						elm$core$Platform$Cmd$none);
				}
			case 'UsernameInput':
				var username = msg.a;
				var _n11 = model.loginState;
				switch (_n11.$) {
					case 'LoginReady':
						var password = _n11.b;
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A2(author$project$Main$LoginReady, username, password)
								}),
							elm$core$Platform$Cmd$none);
					case 'LoginFailure':
						var password = _n11.b;
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A3(author$project$Main$LoginFailure, username, password, '')
								}),
							elm$core$Platform$Cmd$none);
					default:
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'PasswordInput':
				var password = msg.a;
				var _n12 = model.loginState;
				switch (_n12.$) {
					case 'LoginReady':
						var username = _n12.a;
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A2(author$project$Main$LoginReady, username, password)
								}),
							elm$core$Platform$Cmd$none);
					case 'LoginFailure':
						var username = _n12.a;
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A3(author$project$Main$LoginFailure, username, password, '')
								}),
							elm$core$Platform$Cmd$none);
					default:
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'LoginSubmit':
				var _n13 = model.loginState;
				switch (_n13.$) {
					case 'LoginReady':
						var username = _n13.a;
						var password = _n13.b;
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A2(author$project$Main$LoginLoading, username, password)
								}),
							A3(author$project$Commands$doLogin, author$project$Main$GotLoginResp, username, password));
					case 'LoginFailure':
						var username = _n13.a;
						var password = _n13.b;
						return _Utils_Tuple2(
							_Utils_update(
								model,
								{
									loginState: A2(author$project$Main$LoginLoading, username, password)
								}),
							A3(author$project$Commands$doLogin, author$project$Main$GotLoginResp, username, password));
					default:
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'LogoutSubmit':
				var mayReloginAsAnon = msg.a;
				return _Utils_Tuple2(
					model,
					author$project$Commands$doLogout(
						author$project$Main$GotLogoutResp(mayReloginAsAnon)));
			case 'GotoLogin':
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							loginState: A2(author$project$Main$LoginReady, '', '')
						}),
					elm$core$Platform$Cmd$none);
			case 'PingerIn':
				var pingMsg = msg.a;
				return _Utils_Tuple2(
					_Utils_update(
						model,
						{
							serverIsOnline: author$project$Main$pingerMsgToBool(pingMsg)
						}),
					elm$core$Platform$Cmd$none);
			case 'WebsocketIn':
				var event = msg.a;
				var _n14 = author$project$Main$eventType(event);
				switch (_n14) {
					case 'pin':
						var _n15 = model.loginState;
						if (_n15.$ === 'LoginSuccess') {
							var viewState = _n15.a;
							return _Utils_Tuple2(
								model,
								A2(
									elm$core$Platform$Cmd$map,
									author$project$Main$ListMsg,
									author$project$Routes$Ls$doListQueryFromUrl(model.url)));
						} else {
							return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
						}
					case 'fs':
						var _n16 = model.loginState;
						if (_n16.$ === 'LoginSuccess') {
							var viewState = _n16.a;
							return _Utils_Tuple2(
								model,
								elm$core$Platform$Cmd$batch(
									_List_fromArray(
										[
											A2(
											elm$core$Platform$Cmd$map,
											author$project$Main$ListMsg,
											author$project$Routes$Ls$doListQueryFromUrl(model.url)),
											A2(
											elm$core$Platform$Cmd$map,
											author$project$Main$DeletedFilesMsg,
											author$project$Routes$DeletedFiles$reload(viewState.deletedFilesState)),
											A2(
											elm$core$Platform$Cmd$map,
											author$project$Main$CommitsMsg,
											author$project$Routes$Commits$reload(viewState.commitsState))
										])));
						} else {
							return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
						}
					case 'remotes':
						return _Utils_Tuple2(
							model,
							A2(elm$core$Platform$Cmd$map, author$project$Main$RemotesMsg, author$project$Routes$Remotes$reload));
					default:
						return _Utils_Tuple2(model, elm$core$Platform$Cmd$none);
				}
			case 'ListMsg':
				var subMsg = msg.a;
				return A6(
					author$project$Main$withSubUpdate,
					subMsg,
					function ($) {
						return $.listState;
					},
					model,
					author$project$Main$ListMsg,
					author$project$Routes$Ls$update,
					F2(
						function (viewState, newSubModel) {
							return _Utils_update(
								viewState,
								{listState: newSubModel});
						}));
			case 'CommitsMsg':
				var subMsg = msg.a;
				return A6(
					author$project$Main$withSubUpdate,
					subMsg,
					function ($) {
						return $.commitsState;
					},
					model,
					author$project$Main$CommitsMsg,
					author$project$Routes$Commits$update,
					F2(
						function (viewState, newSubModel) {
							return _Utils_update(
								viewState,
								{commitsState: newSubModel});
						}));
			case 'DeletedFilesMsg':
				var subMsg = msg.a;
				return A6(
					author$project$Main$withSubUpdate,
					subMsg,
					function ($) {
						return $.deletedFilesState;
					},
					model,
					author$project$Main$DeletedFilesMsg,
					author$project$Routes$DeletedFiles$update,
					F2(
						function (viewState, newSubModel) {
							return _Utils_update(
								viewState,
								{deletedFilesState: newSubModel});
						}));
			case 'RemotesMsg':
				var subMsg = msg.a;
				return A6(
					author$project$Main$withSubUpdate,
					subMsg,
					function ($) {
						return $.remoteState;
					},
					model,
					author$project$Main$RemotesMsg,
					author$project$Routes$Remotes$update,
					F2(
						function (viewState, newSubModel) {
							return _Utils_update(
								viewState,
								{remoteState: newSubModel});
						}));
			default:
				var subMsg = msg.a;
				return A6(
					author$project$Main$withSubUpdate,
					subMsg,
					function ($) {
						return $.diffState;
					},
					model,
					author$project$Main$DiffMsg,
					author$project$Routes$Diff$update,
					F2(
						function (viewState, newSubModel) {
							return _Utils_update(
								viewState,
								{diffState: newSubModel});
						}));
		}
	});
var author$project$Main$LoginSubmit = {$: 'LoginSubmit'};
var elm$core$String$trim = _String_trim;
var elm$html$Html$span = _VirtualDom_node('span');
var elm$virtual_dom$VirtualDom$text = _VirtualDom_text;
var elm$html$Html$text = elm$virtual_dom$VirtualDom$text;
var elm$html$Html$Attributes$stringProperty = F2(
	function (key, string) {
		return A2(
			_VirtualDom_property,
			key,
			elm$json$Json$Encode$string(string));
	});
var elm$html$Html$Attributes$class = elm$html$Html$Attributes$stringProperty('className');
var elm$html$Html$Attributes$boolProperty = F2(
	function (key, bool) {
		return A2(
			_VirtualDom_property,
			key,
			elm$json$Json$Encode$bool(bool));
	});
var elm$html$Html$Attributes$disabled = elm$html$Html$Attributes$boolProperty('disabled');
var elm$html$Html$Attributes$type_ = elm$html$Html$Attributes$stringProperty('type');
var elm$virtual_dom$VirtualDom$Normal = function (a) {
	return {$: 'Normal', a: a};
};
var elm$virtual_dom$VirtualDom$on = _VirtualDom_on;
var elm$html$Html$Events$on = F2(
	function (event, decoder) {
		return A2(
			elm$virtual_dom$VirtualDom$on,
			event,
			elm$virtual_dom$VirtualDom$Normal(decoder));
	});
var elm$html$Html$Events$onClick = function (msg) {
	return A2(
		elm$html$Html$Events$on,
		'click',
		elm$json$Json$Decode$succeed(msg));
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Attrs = function (a) {
	return {$: 'Attrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Button$attrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Internal$Button$Attrs(attrs_);
};
var elm$html$Html$button = _VirtualDom_node('button');
var elm$core$Maybe$andThen = F2(
	function (callback, maybeValue) {
		if (maybeValue.$ === 'Just') {
			var value = maybeValue.a;
			return callback(value);
		} else {
			return elm$core$Maybe$Nothing;
		}
	});
var elm$core$Tuple$second = function (_n0) {
	var y = _n0.b;
	return y;
};
var elm$html$Html$Attributes$classList = function (classes) {
	return elm$html$Html$Attributes$class(
		A2(
			elm$core$String$join,
			' ',
			A2(
				elm$core$List$map,
				elm$core$Tuple$first,
				A2(elm$core$List$filter, elm$core$Tuple$second, classes))));
};
var rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption = function (size) {
	switch (size.$) {
		case 'XS':
			return elm$core$Maybe$Nothing;
		case 'SM':
			return elm$core$Maybe$Just('sm');
		case 'MD':
			return elm$core$Maybe$Just('md');
		case 'LG':
			return elm$core$Maybe$Just('lg');
		default:
			return elm$core$Maybe$Just('xl');
	}
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$applyModifier = F2(
	function (modifier, options) {
		switch (modifier.$) {
			case 'Size':
				var size = modifier.a;
				return _Utils_update(
					options,
					{
						size: elm$core$Maybe$Just(size)
					});
			case 'Coloring':
				var coloring = modifier.a;
				return _Utils_update(
					options,
					{
						coloring: elm$core$Maybe$Just(coloring)
					});
			case 'Block':
				return _Utils_update(
					options,
					{block: true});
			case 'Disabled':
				var val = modifier.a;
				return _Utils_update(
					options,
					{disabled: val});
			default:
				var attrs = modifier.a;
				return _Utils_update(
					options,
					{
						attributes: _Utils_ap(options.attributes, attrs)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Button$defaultOptions = {attributes: _List_Nil, block: false, coloring: elm$core$Maybe$Nothing, disabled: false, size: elm$core$Maybe$Nothing};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$roleClass = function (role) {
	switch (role.$) {
		case 'Primary':
			return 'primary';
		case 'Secondary':
			return 'secondary';
		case 'Success':
			return 'success';
		case 'Info':
			return 'info';
		case 'Warning':
			return 'warning';
		case 'Danger':
			return 'danger';
		case 'Dark':
			return 'dark';
		case 'Light':
			return 'light';
		default:
			return 'link';
	}
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$buttonAttributes = function (modifiers) {
	var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Internal$Button$applyModifier, rundis$elm_bootstrap$Bootstrap$Internal$Button$defaultOptions, modifiers);
	return _Utils_ap(
		_List_fromArray(
			[
				elm$html$Html$Attributes$classList(
				_List_fromArray(
					[
						_Utils_Tuple2('btn', true),
						_Utils_Tuple2('btn-block', options.block),
						_Utils_Tuple2('disabled', options.disabled)
					])),
				elm$html$Html$Attributes$disabled(options.disabled)
			]),
		_Utils_ap(
			function () {
				var _n0 = A2(elm$core$Maybe$andThen, rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption, options.size);
				if (_n0.$ === 'Just') {
					var s = _n0.a;
					return _List_fromArray(
						[
							elm$html$Html$Attributes$class('btn-' + s)
						]);
				} else {
					return _List_Nil;
				}
			}(),
			_Utils_ap(
				function () {
					var _n1 = options.coloring;
					if (_n1.$ === 'Just') {
						if (_n1.a.$ === 'Roled') {
							var role = _n1.a.a;
							return _List_fromArray(
								[
									elm$html$Html$Attributes$class(
									'btn-' + rundis$elm_bootstrap$Bootstrap$Internal$Button$roleClass(role))
								]);
						} else {
							var role = _n1.a.a;
							return _List_fromArray(
								[
									elm$html$Html$Attributes$class(
									'btn-outline-' + rundis$elm_bootstrap$Bootstrap$Internal$Button$roleClass(role))
								]);
						}
					} else {
						return _List_Nil;
					}
				}(),
				options.attributes)));
};
var rundis$elm_bootstrap$Bootstrap$Button$button = F2(
	function (options, children) {
		return A2(
			elm$html$Html$button,
			rundis$elm_bootstrap$Bootstrap$Internal$Button$buttonAttributes(options),
			children);
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring = function (a) {
	return {$: 'Coloring', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Primary = {$: 'Primary'};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Roled = function (a) {
	return {$: 'Roled', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Button$primary = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Roled(rundis$elm_bootstrap$Bootstrap$Internal$Button$Primary));
var author$project$Main$viewLoginButton = F3(
	function (username, password, isLoading) {
		var loadingClass = isLoading ? 'fa fa-sync fa-sync-animate' : '';
		return A2(
			rundis$elm_bootstrap$Bootstrap$Button$button,
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Button$primary,
					rundis$elm_bootstrap$Bootstrap$Button$attrs(
					_List_fromArray(
						[
							elm$html$Html$Events$onClick(author$project$Main$LoginSubmit),
							elm$html$Html$Attributes$class('login-btn'),
							elm$html$Html$Attributes$type_('submit'),
							elm$html$Html$Attributes$disabled(
							(!elm$core$String$length(
								elm$core$String$trim(username))) || ((!elm$core$String$length(
								elm$core$String$trim(password))) || isLoading))
						]))
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class(loadingClass)
						]),
					_List_Nil),
					elm$html$Html$text(' Log in')
				]));
	});
var author$project$Main$PasswordInput = function (a) {
	return {$: 'PasswordInput', a: a};
};
var author$project$Main$UsernameInput = function (a) {
	return {$: 'UsernameInput', a: a};
};
var elm$html$Html$h2 = _VirtualDom_node('h2');
var rundis$elm_bootstrap$Bootstrap$Form$Input$Attrs = function (a) {
	return {$: 'Attrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$attrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Form$Input$Attrs(attrs_);
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$Id = function (a) {
	return {$: 'Id', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$id = function (id_) {
	return rundis$elm_bootstrap$Bootstrap$Form$Input$Id(id_);
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$Size = function (a) {
	return {$: 'Size', a: a};
};
var rundis$elm_bootstrap$Bootstrap$General$Internal$LG = {$: 'LG'};
var rundis$elm_bootstrap$Bootstrap$Form$Input$large = rundis$elm_bootstrap$Bootstrap$Form$Input$Size(rundis$elm_bootstrap$Bootstrap$General$Internal$LG);
var rundis$elm_bootstrap$Bootstrap$Form$Input$OnInput = function (a) {
	return {$: 'OnInput', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$onInput = function (toMsg) {
	return rundis$elm_bootstrap$Bootstrap$Form$Input$OnInput(toMsg);
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$Password = {$: 'Password'};
var rundis$elm_bootstrap$Bootstrap$Form$Input$Input = function (a) {
	return {$: 'Input', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$Type = function (a) {
	return {$: 'Type', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$create = F2(
	function (tipe, options) {
		return rundis$elm_bootstrap$Bootstrap$Form$Input$Input(
			{
				options: A2(
					elm$core$List$cons,
					rundis$elm_bootstrap$Bootstrap$Form$Input$Type(tipe),
					options)
			});
	});
var elm$html$Html$input = _VirtualDom_node('input');
var elm$core$Maybe$map = F2(
	function (f, maybe) {
		if (maybe.$ === 'Just') {
			var value = maybe.a;
			return elm$core$Maybe$Just(
				f(value));
		} else {
			return elm$core$Maybe$Nothing;
		}
	});
var elm$html$Html$Attributes$id = elm$html$Html$Attributes$stringProperty('id');
var elm$html$Html$Attributes$placeholder = elm$html$Html$Attributes$stringProperty('placeholder');
var elm$html$Html$Attributes$readonly = elm$html$Html$Attributes$boolProperty('readOnly');
var elm$html$Html$Attributes$value = elm$html$Html$Attributes$stringProperty('value');
var elm$html$Html$Events$alwaysStop = function (x) {
	return _Utils_Tuple2(x, true);
};
var elm$virtual_dom$VirtualDom$MayStopPropagation = function (a) {
	return {$: 'MayStopPropagation', a: a};
};
var elm$html$Html$Events$stopPropagationOn = F2(
	function (event, decoder) {
		return A2(
			elm$virtual_dom$VirtualDom$on,
			event,
			elm$virtual_dom$VirtualDom$MayStopPropagation(decoder));
	});
var elm$json$Json$Decode$at = F2(
	function (fields, decoder) {
		return A3(elm$core$List$foldr, elm$json$Json$Decode$field, decoder, fields);
	});
var elm$html$Html$Events$targetValue = A2(
	elm$json$Json$Decode$at,
	_List_fromArray(
		['target', 'value']),
	elm$json$Json$Decode$string);
var elm$html$Html$Events$onInput = function (tagger) {
	return A2(
		elm$html$Html$Events$stopPropagationOn,
		'input',
		A2(
			elm$json$Json$Decode$map,
			elm$html$Html$Events$alwaysStop,
			A2(elm$json$Json$Decode$map, tagger, elm$html$Html$Events$targetValue)));
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$applyModifier = F2(
	function (modifier, options) {
		switch (modifier.$) {
			case 'Size':
				var size_ = modifier.a;
				return _Utils_update(
					options,
					{
						size: elm$core$Maybe$Just(size_)
					});
			case 'Id':
				var id_ = modifier.a;
				return _Utils_update(
					options,
					{
						id: elm$core$Maybe$Just(id_)
					});
			case 'Type':
				var tipe = modifier.a;
				return _Utils_update(
					options,
					{tipe: tipe});
			case 'Disabled':
				var val = modifier.a;
				return _Utils_update(
					options,
					{disabled: val});
			case 'Value':
				var value_ = modifier.a;
				return _Utils_update(
					options,
					{
						value: elm$core$Maybe$Just(value_)
					});
			case 'Placeholder':
				var value_ = modifier.a;
				return _Utils_update(
					options,
					{
						placeholder: elm$core$Maybe$Just(value_)
					});
			case 'OnInput':
				var onInput_ = modifier.a;
				return _Utils_update(
					options,
					{
						onInput: elm$core$Maybe$Just(onInput_)
					});
			case 'Validation':
				var validation_ = modifier.a;
				return _Utils_update(
					options,
					{
						validation: elm$core$Maybe$Just(validation_)
					});
			case 'Readonly':
				var val = modifier.a;
				return _Utils_update(
					options,
					{readonly: val});
			default:
				var attrs_ = modifier.a;
				return _Utils_update(
					options,
					{
						attributes: _Utils_ap(options.attributes, attrs_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Form$Input$Text = {$: 'Text'};
var rundis$elm_bootstrap$Bootstrap$Form$Input$defaultOptions = {attributes: _List_Nil, disabled: false, id: elm$core$Maybe$Nothing, onInput: elm$core$Maybe$Nothing, placeholder: elm$core$Maybe$Nothing, readonly: false, size: elm$core$Maybe$Nothing, tipe: rundis$elm_bootstrap$Bootstrap$Form$Input$Text, validation: elm$core$Maybe$Nothing, value: elm$core$Maybe$Nothing};
var rundis$elm_bootstrap$Bootstrap$Form$Input$sizeAttribute = function (size) {
	return A2(
		elm$core$Maybe$map,
		function (s) {
			return elm$html$Html$Attributes$class('form-control-' + s);
		},
		rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(size));
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$typeAttribute = function (inputType) {
	return elm$html$Html$Attributes$type_(
		function () {
			switch (inputType.$) {
				case 'Text':
					return 'text';
				case 'Password':
					return 'password';
				case 'DatetimeLocal':
					return 'datetime-local';
				case 'Date':
					return 'date';
				case 'Month':
					return 'month';
				case 'Time':
					return 'time';
				case 'Week':
					return 'week';
				case 'Number':
					return 'number';
				case 'Email':
					return 'email';
				case 'Url':
					return 'url';
				case 'Search':
					return 'search';
				case 'Tel':
					return 'tel';
				default:
					return 'color';
			}
		}());
};
var rundis$elm_bootstrap$Bootstrap$Form$FormInternal$validationToString = function (validation) {
	if (validation.$ === 'Success') {
		return 'is-valid';
	} else {
		return 'is-invalid';
	}
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$validationAttribute = function (validation) {
	return elm$html$Html$Attributes$class(
		rundis$elm_bootstrap$Bootstrap$Form$FormInternal$validationToString(validation));
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$toAttributes = function (modifiers) {
	var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Form$Input$applyModifier, rundis$elm_bootstrap$Bootstrap$Form$Input$defaultOptions, modifiers);
	return _Utils_ap(
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('form-control'),
				elm$html$Html$Attributes$disabled(options.disabled),
				elm$html$Html$Attributes$readonly(options.readonly),
				rundis$elm_bootstrap$Bootstrap$Form$Input$typeAttribute(options.tipe)
			]),
		_Utils_ap(
			A2(
				elm$core$List$filterMap,
				elm$core$Basics$identity,
				_List_fromArray(
					[
						A2(elm$core$Maybe$map, elm$html$Html$Attributes$id, options.id),
						A2(elm$core$Maybe$andThen, rundis$elm_bootstrap$Bootstrap$Form$Input$sizeAttribute, options.size),
						A2(elm$core$Maybe$map, elm$html$Html$Attributes$value, options.value),
						A2(elm$core$Maybe$map, elm$html$Html$Attributes$placeholder, options.placeholder),
						A2(elm$core$Maybe$map, elm$html$Html$Events$onInput, options.onInput),
						A2(elm$core$Maybe$map, rundis$elm_bootstrap$Bootstrap$Form$Input$validationAttribute, options.validation)
					])),
			options.attributes));
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$view = function (_n0) {
	var options = _n0.a.options;
	return A2(
		elm$html$Html$input,
		rundis$elm_bootstrap$Bootstrap$Form$Input$toAttributes(options),
		_List_Nil);
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$input = F2(
	function (tipe, options) {
		return rundis$elm_bootstrap$Bootstrap$Form$Input$view(
			A2(rundis$elm_bootstrap$Bootstrap$Form$Input$create, tipe, options));
	});
var rundis$elm_bootstrap$Bootstrap$Form$Input$password = rundis$elm_bootstrap$Bootstrap$Form$Input$input(rundis$elm_bootstrap$Bootstrap$Form$Input$Password);
var rundis$elm_bootstrap$Bootstrap$Form$Input$Placeholder = function (a) {
	return {$: 'Placeholder', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder = function (value_) {
	return rundis$elm_bootstrap$Bootstrap$Form$Input$Placeholder(value_);
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$text = rundis$elm_bootstrap$Bootstrap$Form$Input$input(rundis$elm_bootstrap$Bootstrap$Form$Input$Text);
var rundis$elm_bootstrap$Bootstrap$Form$Input$Value = function (a) {
	return {$: 'Value', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$Input$value = function (value_) {
	return rundis$elm_bootstrap$Bootstrap$Form$Input$Value(value_);
};
var author$project$Main$viewLoginInputs = F2(
	function (username, password) {
		return _List_fromArray(
			[
				A2(
				elm$html$Html$h2,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('login-header')
					]),
				_List_fromArray(
					[
						elm$html$Html$text('Login')
					])),
				rundis$elm_bootstrap$Bootstrap$Form$Input$text(
				_List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Form$Input$id('username-input'),
						rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('login-input')
							])),
						rundis$elm_bootstrap$Bootstrap$Form$Input$large,
						rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Username'),
						rundis$elm_bootstrap$Bootstrap$Form$Input$value(username),
						rundis$elm_bootstrap$Bootstrap$Form$Input$onInput(author$project$Main$UsernameInput)
					])),
				rundis$elm_bootstrap$Bootstrap$Form$Input$password(
				_List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Form$Input$id('password-input'),
						rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('login-input')
							])),
						rundis$elm_bootstrap$Bootstrap$Form$Input$large,
						rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Password'),
						rundis$elm_bootstrap$Bootstrap$Form$Input$value(password),
						rundis$elm_bootstrap$Bootstrap$Form$Input$onInput(author$project$Main$PasswordInput)
					]))
			]);
	});
var elm$html$Html$Events$alwaysPreventDefault = function (msg) {
	return _Utils_Tuple2(msg, true);
};
var elm$virtual_dom$VirtualDom$MayPreventDefault = function (a) {
	return {$: 'MayPreventDefault', a: a};
};
var elm$html$Html$Events$preventDefaultOn = F2(
	function (event, decoder) {
		return A2(
			elm$virtual_dom$VirtualDom$on,
			event,
			elm$virtual_dom$VirtualDom$MayPreventDefault(decoder));
	});
var elm$html$Html$Events$onSubmit = function (msg) {
	return A2(
		elm$html$Html$Events$preventDefaultOn,
		'submit',
		A2(
			elm$json$Json$Decode$map,
			elm$html$Html$Events$alwaysPreventDefault,
			elm$json$Json$Decode$succeed(msg)));
};
var rundis$elm_bootstrap$Bootstrap$Alert$Config = function (a) {
	return {$: 'Config', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Alert$attrs = F2(
	function (attributes, _n0) {
		var configRec = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Alert$Config(
			_Utils_update(
				configRec,
				{attributes: attributes}));
	});
var rundis$elm_bootstrap$Bootstrap$Alert$children = F2(
	function (children_, _n0) {
		var configRec = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Alert$Config(
			_Utils_update(
				configRec,
				{children: children_}));
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Role$Secondary = {$: 'Secondary'};
var rundis$elm_bootstrap$Bootstrap$Alert$config = rundis$elm_bootstrap$Bootstrap$Alert$Config(
	{attributes: _List_Nil, children: _List_Nil, dismissable: elm$core$Maybe$Nothing, role: rundis$elm_bootstrap$Bootstrap$Internal$Role$Secondary, visibility: rundis$elm_bootstrap$Bootstrap$Alert$Shown, withAnimation: false});
var rundis$elm_bootstrap$Bootstrap$Alert$role = F2(
	function (role_, _n0) {
		var configRec = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Alert$Config(
			_Utils_update(
				configRec,
				{role: role_}));
	});
var elm$html$Html$div = _VirtualDom_node('div');
var elm$virtual_dom$VirtualDom$attribute = F2(
	function (key, value) {
		return A2(
			_VirtualDom_attribute,
			_VirtualDom_noOnOrFormAction(key),
			_VirtualDom_noJavaScriptOrHtmlUri(value));
	});
var elm$html$Html$Attributes$attribute = elm$virtual_dom$VirtualDom$attribute;
var rundis$elm_bootstrap$Bootstrap$Alert$StartClose = {$: 'StartClose'};
var rundis$elm_bootstrap$Bootstrap$Alert$clickHandler = F2(
	function (visibility, configRec) {
		var handleClick = F2(
			function (viz, toMsg) {
				return elm$html$Html$Events$onClick(
					toMsg(viz));
			});
		var _n0 = configRec.dismissable;
		if (_n0.$ === 'Just') {
			var dismissMsg = _n0.a;
			return _List_fromArray(
				[
					configRec.withAnimation ? A2(handleClick, rundis$elm_bootstrap$Bootstrap$Alert$StartClose, dismissMsg) : A2(handleClick, rundis$elm_bootstrap$Bootstrap$Alert$Closed, dismissMsg)
				]);
		} else {
			return _List_Nil;
		}
	});
var rundis$elm_bootstrap$Bootstrap$Alert$injectButton = F2(
	function (btn, children_) {
		if (children_.b) {
			var head = children_.a;
			var tail = children_.b;
			return A2(
				elm$core$List$cons,
				head,
				A2(elm$core$List$cons, btn, tail));
		} else {
			return _List_fromArray(
				[btn]);
		}
	});
var rundis$elm_bootstrap$Bootstrap$Alert$isDismissable = function (configRec) {
	var _n0 = configRec.dismissable;
	if (_n0.$ === 'Just') {
		return true;
	} else {
		return false;
	}
};
var rundis$elm_bootstrap$Bootstrap$Alert$maybeAddDismissButton = F3(
	function (visibilty, configRec, children_) {
		return rundis$elm_bootstrap$Bootstrap$Alert$isDismissable(configRec) ? A2(
			rundis$elm_bootstrap$Bootstrap$Alert$injectButton,
			A2(
				elm$html$Html$button,
				_Utils_ap(
					_List_fromArray(
						[
							elm$html$Html$Attributes$type_('button'),
							elm$html$Html$Attributes$class('close'),
							A2(elm$html$Html$Attributes$attribute, 'aria-label', 'close')
						]),
					A2(rundis$elm_bootstrap$Bootstrap$Alert$clickHandler, visibilty, configRec)),
				_List_fromArray(
					[
						A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								A2(elm$html$Html$Attributes$attribute, 'aria-hidden', 'true')
							]),
						_List_fromArray(
							[
								elm$html$Html$text('×')
							]))
					])),
			children_) : children_;
	});
var elm$virtual_dom$VirtualDom$style = _VirtualDom_style;
var elm$html$Html$Attributes$style = elm$virtual_dom$VirtualDom$style;
var rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass = F2(
	function (prefix, role) {
		return elm$html$Html$Attributes$class(
			prefix + ('-' + function () {
				switch (role.$) {
					case 'Primary':
						return 'primary';
					case 'Secondary':
						return 'secondary';
					case 'Success':
						return 'success';
					case 'Info':
						return 'info';
					case 'Warning':
						return 'warning';
					case 'Danger':
						return 'danger';
					case 'Light':
						return 'light';
					default:
						return 'dark';
				}
			}()));
	});
var rundis$elm_bootstrap$Bootstrap$Alert$viewAttributes = F2(
	function (visibility, configRec) {
		var visibiltyAttributes = _Utils_eq(visibility, rundis$elm_bootstrap$Bootstrap$Alert$Closed) ? _List_fromArray(
			[
				A2(elm$html$Html$Attributes$style, 'display', 'none')
			]) : _List_Nil;
		var animationAttributes = function () {
			if (configRec.withAnimation) {
				var _n0 = configRec.dismissable;
				if (_n0.$ === 'Just') {
					var dismissMsg = _n0.a;
					return _List_fromArray(
						[
							A2(
							elm$html$Html$Events$on,
							'transitionend',
							elm$json$Json$Decode$succeed(
								dismissMsg(rundis$elm_bootstrap$Bootstrap$Alert$Closed)))
						]);
				} else {
					return _List_Nil;
				}
			} else {
				return _List_Nil;
			}
		}();
		var alertAttributes = _List_fromArray(
			[
				A2(elm$html$Html$Attributes$attribute, 'role', 'alert'),
				elm$html$Html$Attributes$classList(
				_List_fromArray(
					[
						_Utils_Tuple2('alert', true),
						_Utils_Tuple2(
						'alert-dismissible',
						rundis$elm_bootstrap$Bootstrap$Alert$isDismissable(configRec)),
						_Utils_Tuple2('fade', configRec.withAnimation),
						_Utils_Tuple2(
						'show',
						_Utils_eq(visibility, rundis$elm_bootstrap$Bootstrap$Alert$Shown))
					])),
				A2(rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass, 'alert', configRec.role)
			]);
		return elm$core$List$concat(
			_List_fromArray(
				[configRec.attributes, alertAttributes, visibiltyAttributes, animationAttributes]));
	});
var rundis$elm_bootstrap$Bootstrap$Alert$view = F2(
	function (visibility, _n0) {
		var configRec = _n0.a;
		return A2(
			elm$html$Html$div,
			A2(rundis$elm_bootstrap$Bootstrap$Alert$viewAttributes, visibility, configRec),
			A3(rundis$elm_bootstrap$Bootstrap$Alert$maybeAddDismissButton, visibility, configRec, configRec.children));
	});
var rundis$elm_bootstrap$Bootstrap$Alert$simple = F3(
	function (role_, attributes, children_) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Alert$view,
			rundis$elm_bootstrap$Bootstrap$Alert$Shown,
			A2(
				rundis$elm_bootstrap$Bootstrap$Alert$children,
				children_,
				A2(
					rundis$elm_bootstrap$Bootstrap$Alert$attrs,
					attributes,
					A2(rundis$elm_bootstrap$Bootstrap$Alert$role, role_, rundis$elm_bootstrap$Bootstrap$Alert$config))));
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Role$Danger = {$: 'Danger'};
var rundis$elm_bootstrap$Bootstrap$Alert$simpleDanger = rundis$elm_bootstrap$Bootstrap$Alert$simple(rundis$elm_bootstrap$Bootstrap$Internal$Role$Danger);
var elm$html$Html$form = _VirtualDom_node('form');
var rundis$elm_bootstrap$Bootstrap$Form$form = F2(
	function (attributes, children) {
		return A2(elm$html$Html$form, attributes, children);
	});
var rundis$elm_bootstrap$Bootstrap$Form$applyModifier = F2(
	function (modifier, options) {
		var attrs = modifier.a;
		return _Utils_update(
			options,
			{
				attributes: _Utils_ap(options.attributes, attrs)
			});
	});
var rundis$elm_bootstrap$Bootstrap$Form$defaultOptions = {attributes: _List_Nil};
var rundis$elm_bootstrap$Bootstrap$Form$toAttributes = function (modifiers) {
	var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Form$applyModifier, rundis$elm_bootstrap$Bootstrap$Form$defaultOptions, modifiers);
	return _Utils_ap(
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('form-group')
			]),
		options.attributes);
};
var rundis$elm_bootstrap$Bootstrap$Form$group = F2(
	function (options, children) {
		return A2(
			elm$html$Html$div,
			rundis$elm_bootstrap$Bootstrap$Form$toAttributes(options),
			children);
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Column = function (a) {
	return {$: 'Column', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Grid$col = F2(
	function (options, children) {
		return rundis$elm_bootstrap$Bootstrap$Grid$Column(
			{children: children, options: options});
	});
var rundis$elm_bootstrap$Bootstrap$Grid$containerFluid = F2(
	function (attributes, children) {
		return A2(
			elm$html$Html$div,
			_Utils_ap(
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('container-fluid')
					]),
				attributes),
			children);
	});
var elm$virtual_dom$VirtualDom$keyedNode = function (tag) {
	return _VirtualDom_keyedNode(
		_VirtualDom_noScript(tag));
};
var elm$html$Html$Keyed$node = elm$virtual_dom$VirtualDom$keyedNode;
var rundis$elm_bootstrap$Bootstrap$General$Internal$XS = {$: 'XS'};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col = {$: 'Col'};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Width = F2(
	function (screenSize, columnCount) {
		return {columnCount: columnCount, screenSize: screenSize};
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColAlign = F2(
	function (align_, options) {
		var _n0 = align_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						alignXs: elm$core$Maybe$Just(align_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						alignSm: elm$core$Maybe$Just(align_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						alignMd: elm$core$Maybe$Just(align_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						alignLg: elm$core$Maybe$Just(align_)
					});
			default:
				return _Utils_update(
					options,
					{
						alignXl: elm$core$Maybe$Just(align_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColOffset = F2(
	function (offset_, options) {
		var _n0 = offset_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						offsetXs: elm$core$Maybe$Just(offset_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						offsetSm: elm$core$Maybe$Just(offset_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						offsetMd: elm$core$Maybe$Just(offset_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						offsetLg: elm$core$Maybe$Just(offset_)
					});
			default:
				return _Utils_update(
					options,
					{
						offsetXl: elm$core$Maybe$Just(offset_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColOrder = F2(
	function (order_, options) {
		var _n0 = order_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						orderXs: elm$core$Maybe$Just(order_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						orderSm: elm$core$Maybe$Just(order_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						orderMd: elm$core$Maybe$Just(order_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						orderLg: elm$core$Maybe$Just(order_)
					});
			default:
				return _Utils_update(
					options,
					{
						orderXl: elm$core$Maybe$Just(order_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColPull = F2(
	function (pull_, options) {
		var _n0 = pull_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						pullXs: elm$core$Maybe$Just(pull_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						pullSm: elm$core$Maybe$Just(pull_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						pullMd: elm$core$Maybe$Just(pull_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						pullLg: elm$core$Maybe$Just(pull_)
					});
			default:
				return _Utils_update(
					options,
					{
						pullXl: elm$core$Maybe$Just(pull_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColPush = F2(
	function (push_, options) {
		var _n0 = push_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						pushXs: elm$core$Maybe$Just(push_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						pushSm: elm$core$Maybe$Just(push_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						pushMd: elm$core$Maybe$Just(push_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						pushLg: elm$core$Maybe$Just(push_)
					});
			default:
				return _Utils_update(
					options,
					{
						pushXl: elm$core$Maybe$Just(push_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColWidth = F2(
	function (width_, options) {
		var _n0 = width_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						widthXs: elm$core$Maybe$Just(width_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						widthSm: elm$core$Maybe$Just(width_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						widthMd: elm$core$Maybe$Just(width_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						widthLg: elm$core$Maybe$Just(width_)
					});
			default:
				return _Utils_update(
					options,
					{
						widthXl: elm$core$Maybe$Just(width_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColOption = F2(
	function (modifier, options) {
		switch (modifier.$) {
			case 'ColAttrs':
				var attrs = modifier.a;
				return _Utils_update(
					options,
					{
						attributes: _Utils_ap(options.attributes, attrs)
					});
			case 'ColWidth':
				var width_ = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColWidth, width_, options);
			case 'ColOffset':
				var offset_ = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColOffset, offset_, options);
			case 'ColPull':
				var pull_ = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColPull, pull_, options);
			case 'ColPush':
				var push_ = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColPush, push_, options);
			case 'ColOrder':
				var order_ = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColOrder, order_, options);
			case 'ColAlign':
				var align = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColAlign, align, options);
			default:
				var align = modifier.a;
				return _Utils_update(
					options,
					{
						textAlign: elm$core$Maybe$Just(align)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$columnCountOption = function (size) {
	switch (size.$) {
		case 'Col':
			return elm$core$Maybe$Nothing;
		case 'Col1':
			return elm$core$Maybe$Just('1');
		case 'Col2':
			return elm$core$Maybe$Just('2');
		case 'Col3':
			return elm$core$Maybe$Just('3');
		case 'Col4':
			return elm$core$Maybe$Just('4');
		case 'Col5':
			return elm$core$Maybe$Just('5');
		case 'Col6':
			return elm$core$Maybe$Just('6');
		case 'Col7':
			return elm$core$Maybe$Just('7');
		case 'Col8':
			return elm$core$Maybe$Just('8');
		case 'Col9':
			return elm$core$Maybe$Just('9');
		case 'Col10':
			return elm$core$Maybe$Just('10');
		case 'Col11':
			return elm$core$Maybe$Just('11');
		case 'Col12':
			return elm$core$Maybe$Just('12');
		default:
			return elm$core$Maybe$Just('auto');
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$colWidthClass = function (_n0) {
	var screenSize = _n0.screenSize;
	var columnCount = _n0.columnCount;
	return elm$html$Html$Attributes$class(
		'col' + (A2(
			elm$core$Maybe$withDefault,
			'',
			A2(
				elm$core$Maybe$map,
				function (v) {
					return '-' + v;
				},
				rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(screenSize))) + A2(
			elm$core$Maybe$withDefault,
			'',
			A2(
				elm$core$Maybe$map,
				function (v) {
					return '-' + v;
				},
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$columnCountOption(columnCount)))));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$colWidthsToAttributes = function (widths) {
	var width_ = function (w) {
		return A2(elm$core$Maybe$map, rundis$elm_bootstrap$Bootstrap$Grid$Internal$colWidthClass, w);
	};
	return A2(
		elm$core$List$filterMap,
		elm$core$Basics$identity,
		A2(elm$core$List$map, width_, widths));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$defaultColOptions = {alignLg: elm$core$Maybe$Nothing, alignMd: elm$core$Maybe$Nothing, alignSm: elm$core$Maybe$Nothing, alignXl: elm$core$Maybe$Nothing, alignXs: elm$core$Maybe$Nothing, attributes: _List_Nil, offsetLg: elm$core$Maybe$Nothing, offsetMd: elm$core$Maybe$Nothing, offsetSm: elm$core$Maybe$Nothing, offsetXl: elm$core$Maybe$Nothing, offsetXs: elm$core$Maybe$Nothing, orderLg: elm$core$Maybe$Nothing, orderMd: elm$core$Maybe$Nothing, orderSm: elm$core$Maybe$Nothing, orderXl: elm$core$Maybe$Nothing, orderXs: elm$core$Maybe$Nothing, pullLg: elm$core$Maybe$Nothing, pullMd: elm$core$Maybe$Nothing, pullSm: elm$core$Maybe$Nothing, pullXl: elm$core$Maybe$Nothing, pullXs: elm$core$Maybe$Nothing, pushLg: elm$core$Maybe$Nothing, pushMd: elm$core$Maybe$Nothing, pushSm: elm$core$Maybe$Nothing, pushXl: elm$core$Maybe$Nothing, pushXs: elm$core$Maybe$Nothing, textAlign: elm$core$Maybe$Nothing, widthLg: elm$core$Maybe$Nothing, widthMd: elm$core$Maybe$Nothing, widthSm: elm$core$Maybe$Nothing, widthXl: elm$core$Maybe$Nothing, widthXs: elm$core$Maybe$Nothing};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$offsetCountOption = function (size) {
	switch (size.$) {
		case 'Offset0':
			return '0';
		case 'Offset1':
			return '1';
		case 'Offset2':
			return '2';
		case 'Offset3':
			return '3';
		case 'Offset4':
			return '4';
		case 'Offset5':
			return '5';
		case 'Offset6':
			return '6';
		case 'Offset7':
			return '7';
		case 'Offset8':
			return '8';
		case 'Offset9':
			return '9';
		case 'Offset10':
			return '10';
		default:
			return '11';
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$screenSizeToPartialString = function (screenSize) {
	var _n0 = rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(screenSize);
	if (_n0.$ === 'Just') {
		var s = _n0.a;
		return '-' + (s + '-');
	} else {
		return '-';
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$offsetClass = function (_n0) {
	var screenSize = _n0.screenSize;
	var offsetCount = _n0.offsetCount;
	return elm$html$Html$Attributes$class(
		'offset' + (rundis$elm_bootstrap$Bootstrap$Grid$Internal$screenSizeToPartialString(screenSize) + rundis$elm_bootstrap$Bootstrap$Grid$Internal$offsetCountOption(offsetCount)));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$offsetsToAttributes = function (offsets) {
	var offset_ = function (m) {
		return A2(elm$core$Maybe$map, rundis$elm_bootstrap$Bootstrap$Grid$Internal$offsetClass, m);
	};
	return A2(
		elm$core$List$filterMap,
		elm$core$Basics$identity,
		A2(elm$core$List$map, offset_, offsets));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$orderColOption = function (size) {
	switch (size.$) {
		case 'OrderFirst':
			return 'first';
		case 'Order1':
			return '1';
		case 'Order2':
			return '2';
		case 'Order3':
			return '3';
		case 'Order4':
			return '4';
		case 'Order5':
			return '5';
		case 'Order6':
			return '6';
		case 'Order7':
			return '7';
		case 'Order8':
			return '8';
		case 'Order9':
			return '9';
		case 'Order10':
			return '10';
		case 'Order11':
			return '11';
		case 'Order12':
			return '12';
		default:
			return 'last';
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$orderToAttributes = function (orders) {
	var order_ = function (m) {
		if (m.$ === 'Just') {
			var screenSize = m.a.screenSize;
			var moveCount = m.a.moveCount;
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class(
					'order' + (rundis$elm_bootstrap$Bootstrap$Grid$Internal$screenSizeToPartialString(screenSize) + rundis$elm_bootstrap$Bootstrap$Grid$Internal$orderColOption(moveCount))));
		} else {
			return elm$core$Maybe$Nothing;
		}
	};
	return A2(
		elm$core$List$filterMap,
		elm$core$Basics$identity,
		A2(elm$core$List$map, order_, orders));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$moveCountOption = function (size) {
	switch (size.$) {
		case 'Move0':
			return '0';
		case 'Move1':
			return '1';
		case 'Move2':
			return '2';
		case 'Move3':
			return '3';
		case 'Move4':
			return '4';
		case 'Move5':
			return '5';
		case 'Move6':
			return '6';
		case 'Move7':
			return '7';
		case 'Move8':
			return '8';
		case 'Move9':
			return '9';
		case 'Move10':
			return '10';
		case 'Move11':
			return '11';
		default:
			return '12';
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$pullsToAttributes = function (pulls) {
	var pull_ = function (m) {
		if (m.$ === 'Just') {
			var screenSize = m.a.screenSize;
			var moveCount = m.a.moveCount;
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class(
					'pull' + (rundis$elm_bootstrap$Bootstrap$Grid$Internal$screenSizeToPartialString(screenSize) + rundis$elm_bootstrap$Bootstrap$Grid$Internal$moveCountOption(moveCount))));
		} else {
			return elm$core$Maybe$Nothing;
		}
	};
	return A2(
		elm$core$List$filterMap,
		elm$core$Basics$identity,
		A2(elm$core$List$map, pull_, pulls));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$pushesToAttributes = function (pushes) {
	var push_ = function (m) {
		if (m.$ === 'Just') {
			var screenSize = m.a.screenSize;
			var moveCount = m.a.moveCount;
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class(
					'push' + (rundis$elm_bootstrap$Bootstrap$Grid$Internal$screenSizeToPartialString(screenSize) + rundis$elm_bootstrap$Bootstrap$Grid$Internal$moveCountOption(moveCount))));
		} else {
			return elm$core$Maybe$Nothing;
		}
	};
	return A2(
		elm$core$List$filterMap,
		elm$core$Basics$identity,
		A2(elm$core$List$map, push_, pushes));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$verticalAlignOption = function (align) {
	switch (align.$) {
		case 'Top':
			return 'start';
		case 'Middle':
			return 'center';
		default:
			return 'end';
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$vAlignClass = F2(
	function (prefix, _n0) {
		var align = _n0.align;
		var screenSize = _n0.screenSize;
		return elm$html$Html$Attributes$class(
			_Utils_ap(
				prefix,
				_Utils_ap(
					A2(
						elm$core$Maybe$withDefault,
						'',
						A2(
							elm$core$Maybe$map,
							function (v) {
								return v + '-';
							},
							rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(screenSize))),
					rundis$elm_bootstrap$Bootstrap$Grid$Internal$verticalAlignOption(align))));
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$vAlignsToAttributes = F2(
	function (prefix, aligns) {
		var align = function (a) {
			return A2(
				elm$core$Maybe$map,
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$vAlignClass(prefix),
				a);
		};
		return A2(
			elm$core$List$filterMap,
			elm$core$Basics$identity,
			A2(elm$core$List$map, align, aligns));
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Text$textAlignDirOption = function (dir) {
	switch (dir.$) {
		case 'Center':
			return 'center';
		case 'Left':
			return 'left';
		default:
			return 'right';
	}
};
var rundis$elm_bootstrap$Bootstrap$Internal$Text$textAlignClass = function (_n0) {
	var dir = _n0.dir;
	var size = _n0.size;
	return elm$html$Html$Attributes$class(
		'text' + (A2(
			elm$core$Maybe$withDefault,
			'-',
			A2(
				elm$core$Maybe$map,
				function (s) {
					return '-' + (s + '-');
				},
				rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(size))) + rundis$elm_bootstrap$Bootstrap$Internal$Text$textAlignDirOption(dir)));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$colAttributes = function (modifiers) {
	var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyColOption, rundis$elm_bootstrap$Bootstrap$Grid$Internal$defaultColOptions, modifiers);
	var shouldAddDefaultXs = !elm$core$List$length(
		A2(
			elm$core$List$filterMap,
			elm$core$Basics$identity,
			_List_fromArray(
				[options.widthXs, options.widthSm, options.widthMd, options.widthLg, options.widthXl])));
	return _Utils_ap(
		rundis$elm_bootstrap$Bootstrap$Grid$Internal$colWidthsToAttributes(
			_List_fromArray(
				[
					shouldAddDefaultXs ? elm$core$Maybe$Just(
					A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$Width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col)) : options.widthXs,
					options.widthSm,
					options.widthMd,
					options.widthLg,
					options.widthXl
				])),
		_Utils_ap(
			rundis$elm_bootstrap$Bootstrap$Grid$Internal$offsetsToAttributes(
				_List_fromArray(
					[options.offsetXs, options.offsetSm, options.offsetMd, options.offsetLg, options.offsetXl])),
			_Utils_ap(
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$pullsToAttributes(
					_List_fromArray(
						[options.pullXs, options.pullSm, options.pullMd, options.pullLg, options.pullXl])),
				_Utils_ap(
					rundis$elm_bootstrap$Bootstrap$Grid$Internal$pushesToAttributes(
						_List_fromArray(
							[options.pushXs, options.pushSm, options.pushMd, options.pushLg, options.pushXl])),
					_Utils_ap(
						rundis$elm_bootstrap$Bootstrap$Grid$Internal$orderToAttributes(
							_List_fromArray(
								[options.orderXs, options.orderSm, options.orderMd, options.orderLg, options.orderXl])),
						_Utils_ap(
							A2(
								rundis$elm_bootstrap$Bootstrap$Grid$Internal$vAlignsToAttributes,
								'align-self-',
								_List_fromArray(
									[options.alignXs, options.alignSm, options.alignMd, options.alignLg, options.alignXl])),
							_Utils_ap(
								function () {
									var _n0 = options.textAlign;
									if (_n0.$ === 'Just') {
										var a = _n0.a;
										return _List_fromArray(
											[
												rundis$elm_bootstrap$Bootstrap$Internal$Text$textAlignClass(a)
											]);
									} else {
										return _List_Nil;
									}
								}(),
								options.attributes)))))));
};
var rundis$elm_bootstrap$Bootstrap$Grid$renderCol = function (column) {
	switch (column.$) {
		case 'Column':
			var options = column.a.options;
			var children = column.a.children;
			return A2(
				elm$html$Html$div,
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$colAttributes(options),
				children);
		case 'ColBreak':
			var e = column.a;
			return e;
		default:
			var options = column.a.options;
			var children = column.a.children;
			return A3(
				elm$html$Html$Keyed$node,
				'div',
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$colAttributes(options),
				children);
	}
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyRowHAlign = F2(
	function (align, options) {
		var _n0 = align.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						hAlignXs: elm$core$Maybe$Just(align)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						hAlignSm: elm$core$Maybe$Just(align)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						hAlignMd: elm$core$Maybe$Just(align)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						hAlignLg: elm$core$Maybe$Just(align)
					});
			default:
				return _Utils_update(
					options,
					{
						hAlignXl: elm$core$Maybe$Just(align)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyRowVAlign = F2(
	function (align_, options) {
		var _n0 = align_.screenSize;
		switch (_n0.$) {
			case 'XS':
				return _Utils_update(
					options,
					{
						vAlignXs: elm$core$Maybe$Just(align_)
					});
			case 'SM':
				return _Utils_update(
					options,
					{
						vAlignSm: elm$core$Maybe$Just(align_)
					});
			case 'MD':
				return _Utils_update(
					options,
					{
						vAlignMd: elm$core$Maybe$Just(align_)
					});
			case 'LG':
				return _Utils_update(
					options,
					{
						vAlignLg: elm$core$Maybe$Just(align_)
					});
			default:
				return _Utils_update(
					options,
					{
						vAlignXl: elm$core$Maybe$Just(align_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyRowOption = F2(
	function (modifier, options) {
		switch (modifier.$) {
			case 'RowAttrs':
				var attrs = modifier.a;
				return _Utils_update(
					options,
					{
						attributes: _Utils_ap(options.attributes, attrs)
					});
			case 'RowVAlign':
				var align = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyRowVAlign, align, options);
			default:
				var align = modifier.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyRowHAlign, align, options);
		}
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$defaultRowOptions = {attributes: _List_Nil, hAlignLg: elm$core$Maybe$Nothing, hAlignMd: elm$core$Maybe$Nothing, hAlignSm: elm$core$Maybe$Nothing, hAlignXl: elm$core$Maybe$Nothing, hAlignXs: elm$core$Maybe$Nothing, vAlignLg: elm$core$Maybe$Nothing, vAlignMd: elm$core$Maybe$Nothing, vAlignSm: elm$core$Maybe$Nothing, vAlignXl: elm$core$Maybe$Nothing, vAlignXs: elm$core$Maybe$Nothing};
var rundis$elm_bootstrap$Bootstrap$General$Internal$horizontalAlignOption = function (align) {
	switch (align.$) {
		case 'Left':
			return 'start';
		case 'Center':
			return 'center';
		case 'Right':
			return 'end';
		case 'Around':
			return 'around';
		default:
			return 'between';
	}
};
var rundis$elm_bootstrap$Bootstrap$General$Internal$hAlignClass = function (_n0) {
	var align = _n0.align;
	var screenSize = _n0.screenSize;
	return elm$html$Html$Attributes$class(
		'justify-content-' + (A2(
			elm$core$Maybe$withDefault,
			'',
			A2(
				elm$core$Maybe$map,
				function (v) {
					return v + '-';
				},
				rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(screenSize))) + rundis$elm_bootstrap$Bootstrap$General$Internal$horizontalAlignOption(align)));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$hAlignsToAttributes = function (aligns) {
	var align = function (a) {
		return A2(elm$core$Maybe$map, rundis$elm_bootstrap$Bootstrap$General$Internal$hAlignClass, a);
	};
	return A2(
		elm$core$List$filterMap,
		elm$core$Basics$identity,
		A2(elm$core$List$map, align, aligns));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$rowAttributes = function (modifiers) {
	var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Grid$Internal$applyRowOption, rundis$elm_bootstrap$Bootstrap$Grid$Internal$defaultRowOptions, modifiers);
	return _Utils_ap(
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('row')
			]),
		_Utils_ap(
			A2(
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$vAlignsToAttributes,
				'align-items-',
				_List_fromArray(
					[options.vAlignXs, options.vAlignSm, options.vAlignMd, options.vAlignLg, options.vAlignXl])),
			_Utils_ap(
				rundis$elm_bootstrap$Bootstrap$Grid$Internal$hAlignsToAttributes(
					_List_fromArray(
						[options.hAlignXs, options.hAlignSm, options.hAlignMd, options.hAlignLg, options.hAlignXl])),
				options.attributes)));
};
var rundis$elm_bootstrap$Bootstrap$Grid$row = F2(
	function (options, cols) {
		return A2(
			elm$html$Html$div,
			rundis$elm_bootstrap$Bootstrap$Grid$Internal$rowAttributes(options),
			A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Grid$renderCol, cols));
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$ColAttrs = function (a) {
	return {$: 'ColAttrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Grid$Internal$ColAttrs(attrs_);
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col8 = {$: 'Col8'};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$ColWidth = function (a) {
	return {$: 'ColWidth', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$width = F2(
	function (size, count) {
		return rundis$elm_bootstrap$Bootstrap$Grid$Internal$ColWidth(
			A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$Width, size, count));
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Col$lg8 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$LG, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col8);
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$TextAlign = function (a) {
	return {$: 'TextAlign', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign = function (align) {
	return rundis$elm_bootstrap$Bootstrap$Grid$Internal$TextAlign(align);
};
var rundis$elm_bootstrap$Bootstrap$Internal$Text$Center = {$: 'Center'};
var rundis$elm_bootstrap$Bootstrap$Text$alignXs = function (dir) {
	return {dir: dir, size: rundis$elm_bootstrap$Bootstrap$General$Internal$XS};
};
var rundis$elm_bootstrap$Bootstrap$Text$alignXsCenter = rundis$elm_bootstrap$Bootstrap$Text$alignXs(rundis$elm_bootstrap$Bootstrap$Internal$Text$Center);
var author$project$Main$viewLoginForm = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
		_List_Nil,
		_List_fromArray(
			[
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$row,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$col,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Grid$Col$lg8,
								rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsCenter),
								rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('login-form')
									]))
							]),
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Form$form,
								_List_fromArray(
									[
										elm$html$Html$Events$onSubmit(author$project$Main$LoginSubmit)
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Form$group,
										_List_Nil,
										function () {
											var _n0 = model.loginState;
											switch (_n0.$) {
												case 'LoginReady':
													var username = _n0.a;
													var password = _n0.b;
													return _Utils_ap(
														A2(author$project$Main$viewLoginInputs, username, password),
														_List_fromArray(
															[
																A3(author$project$Main$viewLoginButton, username, password, false)
															]));
												case 'LoginLoading':
													var username = _n0.a;
													var password = _n0.b;
													return _Utils_ap(
														A2(author$project$Main$viewLoginInputs, username, password),
														_List_fromArray(
															[
																A3(author$project$Main$viewLoginButton, username, password, true)
															]));
												case 'LoginFailure':
													var username = _n0.a;
													var password = _n0.b;
													return _Utils_ap(
														A2(author$project$Main$viewLoginInputs, username, password),
														_List_fromArray(
															[
																A2(
																rundis$elm_bootstrap$Bootstrap$Alert$simpleDanger,
																_List_Nil,
																_List_fromArray(
																	[
																		elm$html$Html$text('Login failed, please try again.')
																	])),
																A3(author$project$Main$viewLoginButton, username, password, false)
															]));
												default:
													return _List_Nil;
											}
										}())
									]))
							]))
					]))
			]));
};
var elm$html$Html$a = _VirtualDom_node('a');
var elm$html$Html$Attributes$href = function (url) {
	return A2(
		elm$html$Html$Attributes$stringProperty,
		'href',
		_VirtualDom_noJavaScriptUri(url));
};
var author$project$Main$viewAppIcon = function (model) {
	return A2(
		elm$html$Html$a,
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('nav-link active'),
				elm$html$Html$Attributes$href('/view')
			]),
		model.serverIsOnline ? _List_fromArray(
			[
				A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fas fa-2x fa-fw logo fa-torii-gate')
					]),
				_List_Nil),
				A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('badge badge-success text-center')
					]),
				_List_fromArray(
					[
						elm$html$Html$text('beta')
					]))
			]) : _List_fromArray(
			[
				A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fas fa-2x fa-fw logo logo-failure fa-torii-gate')
					]),
				_List_Nil),
				A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('badge badge-danger text-center')
					]),
				_List_fromArray(
					[
						elm$html$Html$text('offline')
					]))
			]));
};
var author$project$Routes$Commits$CheckoutClicked = function (a) {
	return {$: 'CheckoutClicked', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Danger = {$: 'Danger'};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Outlined = function (a) {
	return {$: 'Outlined', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Button$outlineDanger = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Outlined(rundis$elm_bootstrap$Bootstrap$Internal$Button$Danger));
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col1 = {$: 'Col1'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs1 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col1);
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col3 = {$: 'Col3'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs3 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col3);
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs8 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col8);
var elm$html$Html$li = _VirtualDom_node('li');
var rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$Item = function (a) {
	return {$: 'Item', a: a};
};
var rundis$elm_bootstrap$Bootstrap$ListGroup$li = F2(
	function (options, children) {
		return rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$Item(
			{children: children, itemFn: elm$html$Html$li, options: options});
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Text$Left = {$: 'Left'};
var rundis$elm_bootstrap$Bootstrap$Text$alignXsLeft = rundis$elm_bootstrap$Bootstrap$Text$alignXs(rundis$elm_bootstrap$Bootstrap$Internal$Text$Left);
var rundis$elm_bootstrap$Bootstrap$Internal$Text$Right = {$: 'Right'};
var rundis$elm_bootstrap$Bootstrap$Text$alignXsRight = rundis$elm_bootstrap$Bootstrap$Text$alignXs(rundis$elm_bootstrap$Bootstrap$Internal$Text$Right);
var author$project$Routes$Commits$viewCommit = F2(
	function (model, commit) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$ListGroup$li,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$row,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$col,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Grid$Col$xs1,
									rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsLeft)
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('fas fa-lg fa-save text-xs-right')
										]),
									_List_Nil)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$col,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Grid$Col$xs8,
									rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsLeft)
								]),
							_List_fromArray(
								[
									elm$html$Html$text(commit.msg)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$col,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Grid$Col$xs3,
									rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsRight)
								]),
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Button$button,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Button$outlineDanger,
											rundis$elm_bootstrap$Bootstrap$Button$attrs(
											_List_fromArray(
												[
													elm$html$Html$Events$onClick(
													author$project$Routes$Commits$CheckoutClicked(commit.hash)),
													elm$html$Html$Attributes$disabled(
													((!model.haveStagedChanges) && A2(elm$core$List$member, 'head', commit.tags)) || (!A2(elm$core$List$member, 'fs.edit', model.rights)))
												]))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('Checkout')
										]))
								]))
						]))
				]));
	});
var elm$html$Html$ul = _VirtualDom_node('ul');
var rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$applyModifier = F2(
	function (modifier, options) {
		switch (modifier.$) {
			case 'Roled':
				var role = modifier.a;
				return _Utils_update(
					options,
					{
						role: elm$core$Maybe$Just(role)
					});
			case 'Action':
				return _Utils_update(
					options,
					{action: true});
			case 'Disabled':
				return _Utils_update(
					options,
					{disabled: true});
			case 'Active':
				return _Utils_update(
					options,
					{active: true});
			default:
				var attrs = modifier.a;
				return _Utils_update(
					options,
					{
						attributes: _Utils_ap(options.attributes, attrs)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$defaultOptions = {action: false, active: false, attributes: _List_Nil, disabled: false, role: elm$core$Maybe$Nothing};
var rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$itemAttributes = function (options) {
	return _Utils_ap(
		_List_fromArray(
			[
				elm$html$Html$Attributes$classList(
				_List_fromArray(
					[
						_Utils_Tuple2('list-group-item', true),
						_Utils_Tuple2('disabled', options.disabled),
						_Utils_Tuple2('active', options.active),
						_Utils_Tuple2('list-group-item-action', options.action)
					]))
			]),
		_Utils_ap(
			_List_fromArray(
				[
					elm$html$Html$Attributes$disabled(options.disabled)
				]),
			_Utils_ap(
				A2(
					elm$core$Maybe$withDefault,
					_List_Nil,
					A2(
						elm$core$Maybe$map,
						function (r) {
							return _List_fromArray(
								[
									A2(rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass, 'list-group-item', r)
								]);
						},
						options.role)),
				options.attributes)));
};
var rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$renderItem = function (_n0) {
	var itemFn = _n0.a.itemFn;
	var options = _n0.a.options;
	var children = _n0.a.children;
	return A2(
		itemFn,
		rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$itemAttributes(
			A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$applyModifier, rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$defaultOptions, options)),
		children);
};
var rundis$elm_bootstrap$Bootstrap$ListGroup$ul = function (items) {
	return A2(
		elm$html$Html$ul,
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('list-group')
			]),
		A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$renderItem, items));
};
var author$project$Routes$Commits$viewCommitList = F2(
	function (model, commits) {
		return rundis$elm_bootstrap$Bootstrap$ListGroup$ul(
			A2(
				elm$core$List$map,
				author$project$Routes$Commits$viewCommit(model),
				A2(
					elm$core$List$filter,
					function (c) {
						return elm$core$String$length(c.msg) > 0;
					},
					commits)));
	});
var author$project$Util$iconFromAlertType = function (typ) {
	switch (typ.$) {
		case 'Danger':
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fas fa-xs fa-exclamation-circle')
					]),
				_List_Nil);
		case 'Success':
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fas fa-xs fa-check')
					]),
				_List_Nil);
		default:
			return elm$html$Html$text('');
	}
};
var rundis$elm_bootstrap$Bootstrap$Alert$danger = function (conf) {
	return A2(rundis$elm_bootstrap$Bootstrap$Alert$role, rundis$elm_bootstrap$Bootstrap$Internal$Role$Danger, rundis$elm_bootstrap$Bootstrap$Alert$config);
};
var rundis$elm_bootstrap$Bootstrap$Internal$Role$Info = {$: 'Info'};
var rundis$elm_bootstrap$Bootstrap$Alert$info = function (conf) {
	return A2(rundis$elm_bootstrap$Bootstrap$Alert$role, rundis$elm_bootstrap$Bootstrap$Internal$Role$Info, rundis$elm_bootstrap$Bootstrap$Alert$config);
};
var rundis$elm_bootstrap$Bootstrap$Internal$Role$Success = {$: 'Success'};
var rundis$elm_bootstrap$Bootstrap$Alert$success = function (conf) {
	return A2(rundis$elm_bootstrap$Bootstrap$Alert$role, rundis$elm_bootstrap$Bootstrap$Internal$Role$Success, rundis$elm_bootstrap$Bootstrap$Alert$config);
};
var author$project$Util$visualFromAlertType = function (typ) {
	switch (typ.$) {
		case 'Danger':
			return rundis$elm_bootstrap$Bootstrap$Alert$danger;
		case 'Success':
			return rundis$elm_bootstrap$Bootstrap$Alert$success;
		default:
			return rundis$elm_bootstrap$Bootstrap$Alert$info;
	}
};
var rundis$elm_bootstrap$Bootstrap$Alert$dismissableWithAnimation = F2(
	function (dismissMsg, _n0) {
		var configRec = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Alert$Config(
			_Utils_update(
				configRec,
				{
					dismissable: elm$core$Maybe$Just(dismissMsg),
					withAnimation: true
				}));
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Link = {$: 'Link'};
var rundis$elm_bootstrap$Bootstrap$Button$roleLink = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Roled(rundis$elm_bootstrap$Bootstrap$Internal$Button$Link));
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col10 = {$: 'Col10'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs10 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col10);
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col2 = {$: 'Col2'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs2 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col2);
var author$project$Util$viewAlert = F2(
	function (toMsg, alert) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Alert$view,
			alert.vis,
			A2(
				rundis$elm_bootstrap$Bootstrap$Alert$children,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$row,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs10]),
								_List_fromArray(
									[
										author$project$Util$iconFromAlertType(alert.typ),
										elm$html$Html$text(' ' + alert.message)
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Col$xs2,
										rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsRight)
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Button$button,
										_List_fromArray(
											[
												rundis$elm_bootstrap$Bootstrap$Button$roleLink,
												rundis$elm_bootstrap$Bootstrap$Button$attrs(
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('notification-close-btn'),
														elm$html$Html$Events$onClick(
														toMsg(rundis$elm_bootstrap$Bootstrap$Alert$closed))
													]))
											]),
										_List_fromArray(
											[
												A2(
												elm$html$Html$span,
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('fas fa-xs fa-times')
													]),
												_List_Nil)
											]))
									]))
							]))
					]),
				A2(
					author$project$Util$visualFromAlertType,
					alert.typ,
					A2(rundis$elm_bootstrap$Bootstrap$Alert$dismissableWithAnimation, toMsg, rundis$elm_bootstrap$Bootstrap$Alert$config))));
	});
var elm$html$Html$br = _VirtualDom_node('br');
var elm$html$Html$h4 = _VirtualDom_node('h4');
var rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$LG, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col2);
var rundis$elm_bootstrap$Bootstrap$General$Internal$MD = {$: 'MD'};
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col12 = {$: 'Col12'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$md12 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$MD, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col12);
var author$project$Routes$Commits$viewCommitListContainer = F2(
	function (model, commits) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg8, rundis$elm_bootstrap$Bootstrap$Grid$Col$md12]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted text-center')
								]),
							_List_fromArray(
								[
									elm$html$Html$text('Commits')
								])),
							A2(author$project$Util$viewAlert, author$project$Routes$Commits$AlertMsg, model.alert),
							A2(elm$html$Html$br, _List_Nil, _List_Nil),
							A2(author$project$Routes$Commits$viewCommitList, model, commits),
							A2(elm$html$Html$br, _List_Nil, _List_Nil)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil)
				]));
	});
var author$project$Routes$Commits$SearchInput = function (a) {
	return {$: 'SearchInput', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Config = function (a) {
	return {$: 'Config', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$attrs = F2(
	function (attributes, _n0) {
		var conf = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Config(
			_Utils_update(
				conf,
				{attributes: attributes}));
	});
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$config = function (input_) {
	return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Config(
		{attributes: _List_Nil, input: input_, predecessors: _List_Nil, size: elm$core$Maybe$Nothing, successors: _List_Nil});
};
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Addon = function (a) {
	return {$: 'Addon', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$span = F2(
	function (attributes, children) {
		return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Addon(
			A2(
				elm$html$Html$span,
				A2(
					elm$core$List$cons,
					elm$html$Html$Attributes$class('input-group-text'),
					attributes),
				children));
	});
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$successors = F2(
	function (addons, _n0) {
		var conf = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Config(
			_Utils_update(
				conf,
				{successors: addons}));
	});
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Input = function (a) {
	return {$: 'Input', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$input = F2(
	function (inputFn, options) {
		return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$Input(
			inputFn(options));
	});
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$text = rundis$elm_bootstrap$Bootstrap$Form$InputGroup$input(rundis$elm_bootstrap$Bootstrap$Form$Input$text);
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$sizeAttribute = function (size) {
	return A2(
		elm$core$Maybe$map,
		function (s) {
			return elm$html$Html$Attributes$class('input-group-' + s);
		},
		rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(size));
};
var rundis$elm_bootstrap$Bootstrap$Form$InputGroup$view = function (_n0) {
	var conf = _n0.a;
	var _n1 = conf.input;
	var input_ = _n1.a;
	return A2(
		elm$html$Html$div,
		_Utils_ap(
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('input-group')
				]),
			_Utils_ap(
				A2(
					elm$core$List$filterMap,
					elm$core$Basics$identity,
					_List_fromArray(
						[
							A2(elm$core$Maybe$andThen, rundis$elm_bootstrap$Bootstrap$Form$InputGroup$sizeAttribute, conf.size)
						])),
				conf.attributes)),
		_Utils_ap(
			A2(
				elm$core$List$map,
				function (_n2) {
					var e = _n2.a;
					return A2(
						elm$html$Html$div,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('input-group-prepend')
							]),
						_List_fromArray(
							[e]));
				},
				conf.predecessors),
			_Utils_ap(
				_List_fromArray(
					[input_]),
				A2(
					elm$core$List$map,
					function (_n3) {
						var e = _n3.a;
						return A2(
							elm$html$Html$div,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('input-group-append')
								]),
							_List_fromArray(
								[e]));
					},
					conf.successors))));
};
var author$project$Routes$Commits$viewSearchBox = function (model) {
	return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$view(
		A2(
			rundis$elm_bootstrap$Bootstrap$Form$InputGroup$attrs,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('stylish-input-group input-group')
				]),
			A2(
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$successors,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Form$InputGroup$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('input-group-addon')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$button,
								_List_Nil,
								_List_fromArray(
									[
										A2(
										elm$html$Html$span,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('fas fa-search fa-xs input-group-addon')
											]),
										_List_Nil)
									]))
							]))
					]),
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$config(
					rundis$elm_bootstrap$Bootstrap$Form$InputGroup$text(
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Search'),
								rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onInput(author$project$Routes$Commits$SearchInput),
										elm$html$Html$Attributes$value(model.filter)
									]))
							]))))));
};
var elm$virtual_dom$VirtualDom$lazy = _VirtualDom_lazy;
var elm$html$Html$Lazy$lazy = elm$virtual_dom$VirtualDom$lazy;
var rundis$elm_bootstrap$Bootstrap$Grid$Col$lg12 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$LG, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col12);
var rundis$elm_bootstrap$Bootstrap$General$Internal$XL = {$: 'XL'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xl10 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XL, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col10);
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xl3 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XL, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col3);
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col9 = {$: 'Col9'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xl9 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XL, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col9);
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$RowAttrs = function (a) {
	return {$: 'RowAttrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Grid$Internal$RowAttrs(attrs_);
};
var author$project$Routes$Commits$view = function (model) {
	var _n0 = model.state;
	switch (_n0.$) {
		case 'Loading':
			return elm$html$Html$text('Still loading');
		case 'Failure':
			var err = _n0.a;
			return elm$html$Html$text('Failed to load log: ' + err);
		default:
			var commits = _n0.a;
			return A2(
				rundis$elm_bootstrap$Bootstrap$Grid$row,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$col,
						_List_fromArray(
							[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg12]),
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$id('main-header-row')
											]))
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl9]),
										_List_fromArray(
											[
												elm$html$Html$text('')
											])),
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl3]),
										_List_fromArray(
											[
												A2(elm$html$Html$Lazy$lazy, author$project$Routes$Commits$viewSearchBox, model)
											]))
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$id('main-content-row')
											]))
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl10]),
										_List_fromArray(
											[
												A2(author$project$Routes$Commits$viewCommitListContainer, model, commits)
											]))
									]))
							]))
					]));
	}
};
var author$project$Routes$DeletedFiles$filterEntries = F2(
	function (filter, entries) {
		if (filter === '') {
			return entries;
		} else {
			return A2(
				elm$core$List$filter,
				function (e) {
					return A2(elm$core$String$contains, filter, e.path);
				},
				entries);
		}
	});
var author$project$Routes$DeletedFiles$UndeleteClicked = function (a) {
	return {$: 'UndeleteClicked', a: a};
};
var author$project$Routes$DeletedFiles$viewEntryIcon = function (entry) {
	var _n0 = entry.isDir;
	if (_n0) {
		return A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('fas fa-lg fa-folder text-xs-right file-list-icon')
				]),
			_List_Nil);
	} else {
		return A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('far fa-lg fa-file text-xs-right file-list-icon')
				]),
			_List_Nil);
	}
};
var author$project$Util$monthToInt = function (month) {
	switch (month.$) {
		case 'Jan':
			return 1;
		case 'Feb':
			return 2;
		case 'Mar':
			return 3;
		case 'Apr':
			return 4;
		case 'May':
			return 5;
		case 'Jun':
			return 6;
		case 'Jul':
			return 7;
		case 'Aug':
			return 8;
		case 'Sep':
			return 9;
		case 'Oct':
			return 10;
		case 'Nov':
			return 11;
		default:
			return 12;
	}
};
var elm$core$String$cons = _String_cons;
var elm$core$String$fromChar = function (_char) {
	return A2(elm$core$String$cons, _char, '');
};
var elm$core$Bitwise$shiftRightBy = _Bitwise_shiftRightBy;
var elm$core$String$repeatHelp = F3(
	function (n, chunk, result) {
		return (n <= 0) ? result : A3(
			elm$core$String$repeatHelp,
			n >> 1,
			_Utils_ap(chunk, chunk),
			(!(n & 1)) ? result : _Utils_ap(result, chunk));
	});
var elm$core$String$repeat = F2(
	function (n, chunk) {
		return A3(elm$core$String$repeatHelp, n, chunk, '');
	});
var elm$core$String$padLeft = F3(
	function (n, _char, string) {
		return _Utils_ap(
			A2(
				elm$core$String$repeat,
				n - elm$core$String$length(string),
				elm$core$String$fromChar(_char)),
			string);
	});
var elm$time$Time$flooredDiv = F2(
	function (numerator, denominator) {
		return elm$core$Basics$floor(numerator / denominator);
	});
var elm$time$Time$toAdjustedMinutesHelp = F3(
	function (defaultOffset, posixMinutes, eras) {
		toAdjustedMinutesHelp:
		while (true) {
			if (!eras.b) {
				return posixMinutes + defaultOffset;
			} else {
				var era = eras.a;
				var olderEras = eras.b;
				if (_Utils_cmp(era.start, posixMinutes) < 0) {
					return posixMinutes + era.offset;
				} else {
					var $temp$defaultOffset = defaultOffset,
						$temp$posixMinutes = posixMinutes,
						$temp$eras = olderEras;
					defaultOffset = $temp$defaultOffset;
					posixMinutes = $temp$posixMinutes;
					eras = $temp$eras;
					continue toAdjustedMinutesHelp;
				}
			}
		}
	});
var elm$time$Time$toAdjustedMinutes = F2(
	function (_n0, time) {
		var defaultOffset = _n0.a;
		var eras = _n0.b;
		return A3(
			elm$time$Time$toAdjustedMinutesHelp,
			defaultOffset,
			A2(
				elm$time$Time$flooredDiv,
				elm$time$Time$posixToMillis(time),
				60000),
			eras);
	});
var elm$time$Time$toCivil = function (minutes) {
	var rawDay = A2(elm$time$Time$flooredDiv, minutes, 60 * 24) + 719468;
	var era = (((rawDay >= 0) ? rawDay : (rawDay - 146096)) / 146097) | 0;
	var dayOfEra = rawDay - (era * 146097);
	var yearOfEra = ((((dayOfEra - ((dayOfEra / 1460) | 0)) + ((dayOfEra / 36524) | 0)) - ((dayOfEra / 146096) | 0)) / 365) | 0;
	var dayOfYear = dayOfEra - (((365 * yearOfEra) + ((yearOfEra / 4) | 0)) - ((yearOfEra / 100) | 0));
	var mp = (((5 * dayOfYear) + 2) / 153) | 0;
	var month = mp + ((mp < 10) ? 3 : (-9));
	var year = yearOfEra + (era * 400);
	return {
		day: (dayOfYear - ((((153 * mp) + 2) / 5) | 0)) + 1,
		month: month,
		year: year + ((month <= 2) ? 1 : 0)
	};
};
var elm$time$Time$toDay = F2(
	function (zone, time) {
		return elm$time$Time$toCivil(
			A2(elm$time$Time$toAdjustedMinutes, zone, time)).day;
	});
var elm$time$Time$toHour = F2(
	function (zone, time) {
		return A2(
			elm$core$Basics$modBy,
			24,
			A2(
				elm$time$Time$flooredDiv,
				A2(elm$time$Time$toAdjustedMinutes, zone, time),
				60));
	});
var elm$time$Time$toMinute = F2(
	function (zone, time) {
		return A2(
			elm$core$Basics$modBy,
			60,
			A2(elm$time$Time$toAdjustedMinutes, zone, time));
	});
var elm$time$Time$Apr = {$: 'Apr'};
var elm$time$Time$Aug = {$: 'Aug'};
var elm$time$Time$Dec = {$: 'Dec'};
var elm$time$Time$Feb = {$: 'Feb'};
var elm$time$Time$Jan = {$: 'Jan'};
var elm$time$Time$Jul = {$: 'Jul'};
var elm$time$Time$Jun = {$: 'Jun'};
var elm$time$Time$Mar = {$: 'Mar'};
var elm$time$Time$May = {$: 'May'};
var elm$time$Time$Nov = {$: 'Nov'};
var elm$time$Time$Oct = {$: 'Oct'};
var elm$time$Time$Sep = {$: 'Sep'};
var elm$time$Time$toMonth = F2(
	function (zone, time) {
		var _n0 = elm$time$Time$toCivil(
			A2(elm$time$Time$toAdjustedMinutes, zone, time)).month;
		switch (_n0) {
			case 1:
				return elm$time$Time$Jan;
			case 2:
				return elm$time$Time$Feb;
			case 3:
				return elm$time$Time$Mar;
			case 4:
				return elm$time$Time$Apr;
			case 5:
				return elm$time$Time$May;
			case 6:
				return elm$time$Time$Jun;
			case 7:
				return elm$time$Time$Jul;
			case 8:
				return elm$time$Time$Aug;
			case 9:
				return elm$time$Time$Sep;
			case 10:
				return elm$time$Time$Oct;
			case 11:
				return elm$time$Time$Nov;
			default:
				return elm$time$Time$Dec;
		}
	});
var elm$time$Time$toSecond = F2(
	function (_n0, time) {
		return A2(
			elm$core$Basics$modBy,
			60,
			A2(
				elm$time$Time$flooredDiv,
				elm$time$Time$posixToMillis(time),
				1000));
	});
var elm$time$Time$toYear = F2(
	function (zone, time) {
		return elm$time$Time$toCivil(
			A2(elm$time$Time$toAdjustedMinutes, zone, time)).year;
	});
var author$project$Util$formatLastModified = F2(
	function (z, t) {
		return A2(
			elm$core$String$join,
			' ',
			_List_fromArray(
				[
					A2(
					elm$core$String$join,
					'/',
					_List_fromArray(
						[
							elm$core$String$fromInt(
							A2(elm$time$Time$toDay, z, t)),
							elm$core$String$fromInt(
							author$project$Util$monthToInt(
								A2(elm$time$Time$toMonth, z, t))),
							elm$core$String$fromInt(
							A2(elm$time$Time$toYear, z, t))
						])),
					A2(
					elm$core$String$join,
					':',
					_List_fromArray(
						[
							A3(
							elm$core$String$padLeft,
							2,
							_Utils_chr('0'),
							elm$core$String$fromInt(
								A2(elm$time$Time$toHour, z, t))),
							A3(
							elm$core$String$padLeft,
							2,
							_Utils_chr('0'),
							elm$core$String$fromInt(
								A2(elm$time$Time$toMinute, z, t))),
							A3(
							elm$core$String$padLeft,
							2,
							_Utils_chr('0'),
							elm$core$String$fromInt(
								A2(elm$time$Time$toSecond, z, t)))
						]))
				]));
	});
var basti1302$elm_human_readable_filesize$Filesize$Base10 = {$: 'Base10'};
var basti1302$elm_human_readable_filesize$Filesize$defaultSettings = {decimalPlaces: 2, decimalSeparator: '.', units: basti1302$elm_human_readable_filesize$Filesize$Base10};
var basti1302$elm_human_readable_filesize$Filesize$base10UnitList = _List_fromArray(
	[
		{abbreviation: 'B', minimumSize: 1},
		{abbreviation: 'kB', minimumSize: 1000},
		{abbreviation: 'MB', minimumSize: 1000000},
		{abbreviation: 'GB', minimumSize: 1000000000},
		{abbreviation: 'TB', minimumSize: 1000000000000},
		{abbreviation: 'PB', minimumSize: 1000000000000000},
		{abbreviation: 'EB', minimumSize: 1000000000000000000}
	]);
var basti1302$elm_human_readable_filesize$Filesize$base2UnitList = _List_fromArray(
	[
		{abbreviation: 'B', minimumSize: 1},
		{abbreviation: 'KiB', minimumSize: 1024},
		{abbreviation: 'MiB', minimumSize: 1048576},
		{abbreviation: 'GiB', minimumSize: 1073741824},
		{abbreviation: 'TiB', minimumSize: 1099511627776},
		{abbreviation: 'PiB', minimumSize: 1125899906842624}
	]);
var basti1302$elm_human_readable_filesize$Filesize$getUnitDefinitionList = function (units) {
	if (units.$ === 'Base10') {
		return basti1302$elm_human_readable_filesize$Filesize$base10UnitList;
	} else {
		return basti1302$elm_human_readable_filesize$Filesize$base2UnitList;
	}
};
var basti1302$elm_human_readable_filesize$Filesize$decimalSeparatorRegex = A2(
	elm$core$Maybe$withDefault,
	elm$regex$Regex$never,
	elm$regex$Regex$fromString('\\.'));
var basti1302$elm_human_readable_filesize$Filesize$removeTrailingZeroesRegex = A2(
	elm$core$Maybe$withDefault,
	elm$regex$Regex$never,
	elm$regex$Regex$fromString('^(\\d+\\.[^0]*)(0+)$'));
var elm$core$String$dropRight = F2(
	function (n, string) {
		return (n < 1) ? string : A3(elm$core$String$slice, 0, -n, string);
	});
var elm$core$String$endsWith = _String_endsWith;
var elm$core$String$foldr = _String_foldr;
var elm$core$String$toList = function (string) {
	return A3(elm$core$String$foldr, elm$core$List$cons, _List_Nil, string);
};
var elm$core$Basics$abs = function (n) {
	return (n < 0) ? (-n) : n;
};
var elm$core$Basics$isInfinite = _Basics_isInfinite;
var elm$core$Basics$isNaN = _Basics_isNaN;
var elm$core$String$fromFloat = _String_fromNumber;
var elm$core$String$padRight = F3(
	function (n, _char, string) {
		return _Utils_ap(
			string,
			A2(
				elm$core$String$repeat,
				n - elm$core$String$length(string),
				elm$core$String$fromChar(_char)));
	});
var elm$core$String$reverse = _String_reverse;
var myrho$elm_round$Round$addSign = F2(
	function (signed, str) {
		var isNotZero = A2(
			elm$core$List$any,
			function (c) {
				return (!_Utils_eq(
					c,
					_Utils_chr('0'))) && (!_Utils_eq(
					c,
					_Utils_chr('.')));
			},
			elm$core$String$toList(str));
		return _Utils_ap(
			(signed && isNotZero) ? '-' : '',
			str);
	});
var elm$core$Char$fromCode = _Char_fromCode;
var myrho$elm_round$Round$increaseNum = function (_n0) {
	var head = _n0.a;
	var tail = _n0.b;
	if (_Utils_eq(
		head,
		_Utils_chr('9'))) {
		var _n1 = elm$core$String$uncons(tail);
		if (_n1.$ === 'Nothing') {
			return '01';
		} else {
			var headtail = _n1.a;
			return A2(
				elm$core$String$cons,
				_Utils_chr('0'),
				myrho$elm_round$Round$increaseNum(headtail));
		}
	} else {
		var c = elm$core$Char$toCode(head);
		return ((c >= 48) && (c < 57)) ? A2(
			elm$core$String$cons,
			elm$core$Char$fromCode(c + 1),
			tail) : '0';
	}
};
var myrho$elm_round$Round$splitComma = function (str) {
	var _n0 = A2(elm$core$String$split, '.', str);
	if (_n0.b) {
		if (_n0.b.b) {
			var before = _n0.a;
			var _n1 = _n0.b;
			var after = _n1.a;
			return _Utils_Tuple2(before, after);
		} else {
			var before = _n0.a;
			return _Utils_Tuple2(before, '0');
		}
	} else {
		return _Utils_Tuple2('0', '0');
	}
};
var elm$core$Tuple$mapFirst = F2(
	function (func, _n0) {
		var x = _n0.a;
		var y = _n0.b;
		return _Utils_Tuple2(
			func(x),
			y);
	});
var myrho$elm_round$Round$toDecimal = function (fl) {
	var _n0 = A2(
		elm$core$String$split,
		'e',
		elm$core$String$fromFloat(
			elm$core$Basics$abs(fl)));
	if (_n0.b) {
		if (_n0.b.b) {
			var num = _n0.a;
			var _n1 = _n0.b;
			var exp = _n1.a;
			var e = A2(
				elm$core$Maybe$withDefault,
				0,
				elm$core$String$toInt(
					A2(elm$core$String$startsWith, '+', exp) ? A2(elm$core$String$dropLeft, 1, exp) : exp));
			var _n2 = myrho$elm_round$Round$splitComma(num);
			var before = _n2.a;
			var after = _n2.b;
			var total = _Utils_ap(before, after);
			var zeroed = (e < 0) ? A2(
				elm$core$Maybe$withDefault,
				'0',
				A2(
					elm$core$Maybe$map,
					function (_n3) {
						var a = _n3.a;
						var b = _n3.b;
						return a + ('.' + b);
					},
					A2(
						elm$core$Maybe$map,
						elm$core$Tuple$mapFirst(elm$core$String$fromChar),
						elm$core$String$uncons(
							_Utils_ap(
								A2(
									elm$core$String$repeat,
									elm$core$Basics$abs(e),
									'0'),
								total))))) : A3(
				elm$core$String$padRight,
				e + 1,
				_Utils_chr('0'),
				total);
			return _Utils_ap(
				(fl < 0) ? '-' : '',
				zeroed);
		} else {
			var num = _n0.a;
			return _Utils_ap(
				(fl < 0) ? '-' : '',
				num);
		}
	} else {
		return '';
	}
};
var myrho$elm_round$Round$roundFun = F3(
	function (functor, s, fl) {
		if (elm$core$Basics$isInfinite(fl) || elm$core$Basics$isNaN(fl)) {
			return elm$core$String$fromFloat(fl);
		} else {
			var signed = fl < 0;
			var _n0 = myrho$elm_round$Round$splitComma(
				myrho$elm_round$Round$toDecimal(
					elm$core$Basics$abs(fl)));
			var before = _n0.a;
			var after = _n0.b;
			var r = elm$core$String$length(before) + s;
			var normalized = _Utils_ap(
				A2(elm$core$String$repeat, (-r) + 1, '0'),
				A3(
					elm$core$String$padRight,
					r,
					_Utils_chr('0'),
					_Utils_ap(before, after)));
			var totalLen = elm$core$String$length(normalized);
			var roundDigitIndex = A2(elm$core$Basics$max, 1, r);
			var increase = A2(
				functor,
				signed,
				A3(elm$core$String$slice, roundDigitIndex, totalLen, normalized));
			var remains = A3(elm$core$String$slice, 0, roundDigitIndex, normalized);
			var num = increase ? elm$core$String$reverse(
				A2(
					elm$core$Maybe$withDefault,
					'1',
					A2(
						elm$core$Maybe$map,
						myrho$elm_round$Round$increaseNum,
						elm$core$String$uncons(
							elm$core$String$reverse(remains))))) : remains;
			var numLen = elm$core$String$length(num);
			var numZeroed = (num === '0') ? num : ((s <= 0) ? _Utils_ap(
				num,
				A2(
					elm$core$String$repeat,
					elm$core$Basics$abs(s),
					'0')) : ((_Utils_cmp(
				s,
				elm$core$String$length(after)) < 0) ? (A3(elm$core$String$slice, 0, numLen - s, num) + ('.' + A3(elm$core$String$slice, numLen - s, numLen, num))) : _Utils_ap(
				before + '.',
				A3(
					elm$core$String$padRight,
					s,
					_Utils_chr('0'),
					after))));
			return A2(myrho$elm_round$Round$addSign, signed, numZeroed);
		}
	});
var myrho$elm_round$Round$floor = myrho$elm_round$Round$roundFun(
	F2(
		function (signed, str) {
			var _n0 = elm$core$String$uncons(str);
			if (_n0.$ === 'Nothing') {
				return false;
			} else {
				if ('0' === _n0.a.a.valueOf()) {
					var _n1 = _n0.a;
					var rest = _n1.b;
					return signed && A2(
						elm$core$List$any,
						elm$core$Basics$neq(
							_Utils_chr('0')),
						elm$core$String$toList(rest));
				} else {
					return signed;
				}
			}
		}));
var basti1302$elm_human_readable_filesize$Filesize$roundToDecimalPlaces = F2(
	function (settings, num) {
		var rounded = A2(myrho$elm_round$Round$floor, settings.decimalPlaces, num);
		var withoutTrailingZeroes = A4(
			elm$regex$Regex$replaceAtMost,
			1,
			basti1302$elm_human_readable_filesize$Filesize$removeTrailingZeroesRegex,
			function (_n1) {
				var submatches = _n1.submatches;
				return A2(
					elm$core$String$join,
					'',
					A2(
						elm$core$List$map,
						elm$core$Maybe$withDefault(''),
						A2(elm$core$List$take, 1, submatches)));
			},
			rounded);
		var withoutTrailingDot = A2(elm$core$String$endsWith, '.', withoutTrailingZeroes) ? A2(elm$core$String$dropRight, 1, withoutTrailingZeroes) : withoutTrailingZeroes;
		return (settings.decimalSeparator === '.') ? withoutTrailingDot : A4(
			elm$regex$Regex$replaceAtMost,
			1,
			basti1302$elm_human_readable_filesize$Filesize$decimalSeparatorRegex,
			function (_n0) {
				return settings.decimalSeparator;
			},
			withoutTrailingDot);
	});
var basti1302$elm_human_readable_filesize$Filesize$unknownUnit = {abbreviation: '?', minimumSize: 1};
var basti1302$elm_human_readable_filesize$Filesize$formatWithSplit = F2(
	function (settings, num) {
		if (!num) {
			return _Utils_Tuple2('0', 'B');
		} else {
			var unitDefinitionList = basti1302$elm_human_readable_filesize$Filesize$getUnitDefinitionList(settings.units);
			var _n0 = (num < 0) ? _Utils_Tuple2(-num, '-') : _Utils_Tuple2(num, '');
			var num2 = _n0.a;
			var negativePrefix = _n0.b;
			var unitDefinition = A2(
				elm$core$Maybe$withDefault,
				basti1302$elm_human_readable_filesize$Filesize$unknownUnit,
				elm$core$List$head(
					elm$core$List$reverse(
						A2(
							elm$core$List$filter,
							function (unitDef) {
								return _Utils_cmp(num2, unitDef.minimumSize) > -1;
							},
							unitDefinitionList))));
			var formattedNumber = A2(basti1302$elm_human_readable_filesize$Filesize$roundToDecimalPlaces, settings, num2 / unitDefinition.minimumSize);
			return _Utils_Tuple2(
				_Utils_ap(negativePrefix, formattedNumber),
				unitDefinition.abbreviation);
		}
	});
var basti1302$elm_human_readable_filesize$Filesize$format = function (num) {
	var _n0 = A2(basti1302$elm_human_readable_filesize$Filesize$formatWithSplit, basti1302$elm_human_readable_filesize$Filesize$defaultSettings, num);
	var size = _n0.a;
	var unit = _n0.b;
	return size + (' ' + unit);
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Success = {$: 'Success'};
var rundis$elm_bootstrap$Bootstrap$Button$outlineSuccess = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Outlined(rundis$elm_bootstrap$Bootstrap$Internal$Button$Success));
var rundis$elm_bootstrap$Bootstrap$Table$Td = function (a) {
	return {$: 'Td', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$td = F2(
	function (options, children) {
		return rundis$elm_bootstrap$Bootstrap$Table$Td(
			{children: children, options: options});
	});
var rundis$elm_bootstrap$Bootstrap$Table$Row = function (a) {
	return {$: 'Row', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$tr = F2(
	function (options, cells) {
		return rundis$elm_bootstrap$Bootstrap$Table$Row(
			{cells: cells, options: options});
	});
var author$project$Routes$DeletedFiles$viewDeletedEntry = F2(
	function (model, entry) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Table$tr,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							author$project$Routes$DeletedFiles$viewEntryIcon(entry)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text(entry.path)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text(
							A2(author$project$Util$formatLastModified, model.zone, entry.lastModified))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text(
							basti1302$elm_human_readable_filesize$Filesize$format(entry.size))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Button$button,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Button$outlineSuccess,
									rundis$elm_bootstrap$Bootstrap$Button$attrs(
									_List_fromArray(
										[
											elm$html$Html$Events$onClick(
											author$project$Routes$DeletedFiles$UndeleteClicked(entry.path)),
											elm$html$Html$Attributes$disabled(
											!A2(elm$core$List$member, 'fs.edit', model.rights))
										]))
								]),
							_List_fromArray(
								[
									elm$html$Html$text('Undelete')
								]))
						]))
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Table$TableAttr = function (a) {
	return {$: 'TableAttr', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$attr = function (attr_) {
	return rundis$elm_bootstrap$Bootstrap$Table$TableAttr(attr_);
};
var rundis$elm_bootstrap$Bootstrap$Table$CellAttr = function (a) {
	return {$: 'CellAttr', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$cellAttr = function (attr_) {
	return rundis$elm_bootstrap$Bootstrap$Table$CellAttr(attr_);
};
var rundis$elm_bootstrap$Bootstrap$Table$Hover = {$: 'Hover'};
var rundis$elm_bootstrap$Bootstrap$Table$hover = rundis$elm_bootstrap$Bootstrap$Table$Hover;
var elm$html$Html$table = _VirtualDom_node('table');
var rundis$elm_bootstrap$Bootstrap$Table$Inversed = {$: 'Inversed'};
var rundis$elm_bootstrap$Bootstrap$Table$isResponsive = function (option) {
	if (option.$ === 'Responsive') {
		return true;
	} else {
		return false;
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$KeyedTBody = function (a) {
	return {$: 'KeyedTBody', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$TBody = function (a) {
	return {$: 'TBody', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$InversedRow = function (a) {
	return {$: 'InversedRow', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$KeyedRow = function (a) {
	return {$: 'KeyedRow', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$InversedCell = function (a) {
	return {$: 'InversedCell', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$Th = function (a) {
	return {$: 'Th', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$mapInversedCell = function (cell) {
	var inverseOptions = function (options) {
		return A2(
			elm$core$List$map,
			function (opt) {
				if (opt.$ === 'RoledCell') {
					var role = opt.a;
					return rundis$elm_bootstrap$Bootstrap$Table$InversedCell(role);
				} else {
					return opt;
				}
			},
			options);
	};
	if (cell.$ === 'Th') {
		var cellCfg = cell.a;
		return rundis$elm_bootstrap$Bootstrap$Table$Th(
			_Utils_update(
				cellCfg,
				{
					options: inverseOptions(cellCfg.options)
				}));
	} else {
		var cellCfg = cell.a;
		return rundis$elm_bootstrap$Bootstrap$Table$Td(
			_Utils_update(
				cellCfg,
				{
					options: inverseOptions(cellCfg.options)
				}));
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$mapInversedRow = function (row) {
	var inversedOptions = function (options) {
		return A2(
			elm$core$List$map,
			function (opt) {
				if (opt.$ === 'RoledRow') {
					var role = opt.a;
					return rundis$elm_bootstrap$Bootstrap$Table$InversedRow(role);
				} else {
					return opt;
				}
			},
			options);
	};
	if (row.$ === 'Row') {
		var options = row.a.options;
		var cells = row.a.cells;
		return rundis$elm_bootstrap$Bootstrap$Table$Row(
			{
				cells: A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$mapInversedCell, cells),
				options: inversedOptions(options)
			});
	} else {
		var options = row.a.options;
		var cells = row.a.cells;
		return rundis$elm_bootstrap$Bootstrap$Table$KeyedRow(
			{
				cells: A2(
					elm$core$List$map,
					function (_n1) {
						var key = _n1.a;
						var cell = _n1.b;
						return _Utils_Tuple2(
							key,
							rundis$elm_bootstrap$Bootstrap$Table$mapInversedCell(cell));
					},
					cells),
				options: inversedOptions(options)
			});
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$maybeMapInversedTBody = F2(
	function (isTableInversed, tbody_) {
		var _n0 = _Utils_Tuple2(isTableInversed, tbody_);
		if (!_n0.a) {
			return tbody_;
		} else {
			if (_n0.b.$ === 'TBody') {
				var body = _n0.b.a;
				return rundis$elm_bootstrap$Bootstrap$Table$TBody(
					_Utils_update(
						body,
						{
							rows: A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$mapInversedRow, body.rows)
						}));
			} else {
				var keyedBody = _n0.b.a;
				return rundis$elm_bootstrap$Bootstrap$Table$KeyedTBody(
					_Utils_update(
						keyedBody,
						{
							rows: A2(
								elm$core$List$map,
								function (_n1) {
									var key = _n1.a;
									var row = _n1.b;
									return _Utils_Tuple2(
										key,
										rundis$elm_bootstrap$Bootstrap$Table$mapInversedRow(row));
								},
								keyedBody.rows)
						}));
			}
		}
	});
var rundis$elm_bootstrap$Bootstrap$Table$InversedHead = {$: 'InversedHead'};
var rundis$elm_bootstrap$Bootstrap$Table$THead = function (a) {
	return {$: 'THead', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$maybeMapInversedTHead = F2(
	function (isTableInversed, _n0) {
		var thead_ = _n0.a;
		var isHeadInversed = A2(
			elm$core$List$any,
			function (opt) {
				return _Utils_eq(opt, rundis$elm_bootstrap$Bootstrap$Table$InversedHead);
			},
			thead_.options);
		return rundis$elm_bootstrap$Bootstrap$Table$THead(
			(isTableInversed || isHeadInversed) ? _Utils_update(
				thead_,
				{
					rows: A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$mapInversedRow, thead_.rows)
				}) : thead_);
	});
var rundis$elm_bootstrap$Bootstrap$Table$maybeWrapResponsive = F2(
	function (options, table_) {
		var responsiveClass = elm$html$Html$Attributes$class(
			'table-responsive' + A2(
				elm$core$Maybe$withDefault,
				'',
				A2(
					elm$core$Maybe$map,
					function (v) {
						return '-' + v;
					},
					A2(
						elm$core$Maybe$andThen,
						rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption,
						A2(
							elm$core$Maybe$andThen,
							function (opt) {
								if (opt.$ === 'Responsive') {
									var val = opt.a;
									return val;
								} else {
									return elm$core$Maybe$Nothing;
								}
							},
							elm$core$List$head(
								A2(elm$core$List$filter, rundis$elm_bootstrap$Bootstrap$Table$isResponsive, options)))))));
		return A2(elm$core$List$any, rundis$elm_bootstrap$Bootstrap$Table$isResponsive, options) ? A2(
			elm$html$Html$div,
			_List_fromArray(
				[responsiveClass]),
			_List_fromArray(
				[table_])) : table_;
	});
var elm$html$Html$tbody = _VirtualDom_node('tbody');
var elm$html$Html$Attributes$scope = elm$html$Html$Attributes$stringProperty('scope');
var rundis$elm_bootstrap$Bootstrap$Table$addScopeIfTh = function (cell) {
	if (cell.$ === 'Th') {
		var cellConfig = cell.a;
		return rundis$elm_bootstrap$Bootstrap$Table$Th(
			_Utils_update(
				cellConfig,
				{
					options: A2(
						elm$core$List$cons,
						rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Attributes$scope('row')),
						cellConfig.options)
				}));
	} else {
		return cell;
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$maybeAddScopeToFirstCell = function (row) {
	if (row.$ === 'Row') {
		var options = row.a.options;
		var cells = row.a.cells;
		if (!cells.b) {
			return row;
		} else {
			var first = cells.a;
			var rest = cells.b;
			return rundis$elm_bootstrap$Bootstrap$Table$Row(
				{
					cells: A2(
						elm$core$List$cons,
						rundis$elm_bootstrap$Bootstrap$Table$addScopeIfTh(first),
						rest),
					options: options
				});
		}
	} else {
		var options = row.a.options;
		var cells = row.a.cells;
		if (!cells.b) {
			return row;
		} else {
			var _n3 = cells.a;
			var firstKey = _n3.a;
			var first = _n3.b;
			var rest = cells.b;
			return rundis$elm_bootstrap$Bootstrap$Table$KeyedRow(
				{
					cells: A2(
						elm$core$List$cons,
						_Utils_Tuple2(
							firstKey,
							rundis$elm_bootstrap$Bootstrap$Table$addScopeIfTh(first)),
						rest),
					options: options
				});
		}
	}
};
var elm$html$Html$tr = _VirtualDom_node('tr');
var elm$html$Html$td = _VirtualDom_node('td');
var elm$html$Html$th = _VirtualDom_node('th');
var rundis$elm_bootstrap$Bootstrap$Table$cellAttribute = function (option) {
	switch (option.$) {
		case 'RoledCell':
			if (option.a.$ === 'Roled') {
				var role = option.a.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass, 'table', role);
			} else {
				var _n1 = option.a;
				return elm$html$Html$Attributes$class('table-active');
			}
		case 'InversedCell':
			if (option.a.$ === 'Roled') {
				var role = option.a.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass, 'bg-', role);
			} else {
				var _n2 = option.a;
				return elm$html$Html$Attributes$class('bg-active');
			}
		default:
			var attr_ = option.a;
			return attr_;
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$cellAttributes = function (options) {
	return A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$cellAttribute, options);
};
var rundis$elm_bootstrap$Bootstrap$Table$renderCell = function (cell) {
	if (cell.$ === 'Td') {
		var options = cell.a.options;
		var children = cell.a.children;
		return A2(
			elm$html$Html$td,
			rundis$elm_bootstrap$Bootstrap$Table$cellAttributes(options),
			children);
	} else {
		var options = cell.a.options;
		var children = cell.a.children;
		return A2(
			elm$html$Html$th,
			rundis$elm_bootstrap$Bootstrap$Table$cellAttributes(options),
			children);
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$rowClass = function (option) {
	switch (option.$) {
		case 'RoledRow':
			if (option.a.$ === 'Roled') {
				var role_ = option.a.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass, 'table', role_);
			} else {
				var _n1 = option.a;
				return elm$html$Html$Attributes$class('table-active');
			}
		case 'InversedRow':
			if (option.a.$ === 'Roled') {
				var role_ = option.a.a;
				return A2(rundis$elm_bootstrap$Bootstrap$Internal$Role$toClass, 'bg', role_);
			} else {
				var _n2 = option.a;
				return elm$html$Html$Attributes$class('bg-active');
			}
		default:
			var attr_ = option.a;
			return attr_;
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$rowAttributes = function (options) {
	return A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$rowClass, options);
};
var rundis$elm_bootstrap$Bootstrap$Table$renderRow = function (row) {
	if (row.$ === 'Row') {
		var options = row.a.options;
		var cells = row.a.cells;
		return A2(
			elm$html$Html$tr,
			rundis$elm_bootstrap$Bootstrap$Table$rowAttributes(options),
			A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$renderCell, cells));
	} else {
		var options = row.a.options;
		var cells = row.a.cells;
		return A3(
			elm$html$Html$Keyed$node,
			'tr',
			rundis$elm_bootstrap$Bootstrap$Table$rowAttributes(options),
			A2(
				elm$core$List$map,
				function (_n1) {
					var key = _n1.a;
					var cell = _n1.b;
					return _Utils_Tuple2(
						key,
						rundis$elm_bootstrap$Bootstrap$Table$renderCell(cell));
				},
				cells));
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$renderTBody = function (body) {
	if (body.$ === 'TBody') {
		var attributes = body.a.attributes;
		var rows = body.a.rows;
		return A2(
			elm$html$Html$tbody,
			attributes,
			A2(
				elm$core$List$map,
				function (row) {
					return rundis$elm_bootstrap$Bootstrap$Table$renderRow(
						rundis$elm_bootstrap$Bootstrap$Table$maybeAddScopeToFirstCell(row));
				},
				rows));
	} else {
		var attributes = body.a.attributes;
		var rows = body.a.rows;
		return A3(
			elm$html$Html$Keyed$node,
			'tbody',
			attributes,
			A2(
				elm$core$List$map,
				function (_n1) {
					var key = _n1.a;
					var row = _n1.b;
					return _Utils_Tuple2(
						key,
						rundis$elm_bootstrap$Bootstrap$Table$renderRow(
							rundis$elm_bootstrap$Bootstrap$Table$maybeAddScopeToFirstCell(row)));
				},
				rows));
	}
};
var elm$html$Html$thead = _VirtualDom_node('thead');
var rundis$elm_bootstrap$Bootstrap$Table$theadAttribute = function (option) {
	switch (option.$) {
		case 'InversedHead':
			return elm$html$Html$Attributes$class('thead-dark');
		case 'DefaultHead':
			return elm$html$Html$Attributes$class('thead-default');
		default:
			var attr_ = option.a;
			return attr_;
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$theadAttributes = function (options) {
	return A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$theadAttribute, options);
};
var rundis$elm_bootstrap$Bootstrap$Table$renderTHead = function (_n0) {
	var options = _n0.a.options;
	var rows = _n0.a.rows;
	return A2(
		elm$html$Html$thead,
		rundis$elm_bootstrap$Bootstrap$Table$theadAttributes(options),
		A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$renderRow, rows));
};
var rundis$elm_bootstrap$Bootstrap$Table$tableClass = function (option) {
	switch (option.$) {
		case 'Inversed':
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class('table-dark'));
		case 'Striped':
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class('table-striped'));
		case 'Bordered':
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class('table-bordered'));
		case 'Hover':
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class('table-hover'));
		case 'Small':
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class('table-sm'));
		case 'Responsive':
			return elm$core$Maybe$Nothing;
		case 'Reflow':
			return elm$core$Maybe$Just(
				elm$html$Html$Attributes$class('table-reflow'));
		default:
			var attr_ = option.a;
			return elm$core$Maybe$Just(attr_);
	}
};
var rundis$elm_bootstrap$Bootstrap$Table$tableAttributes = function (options) {
	return A2(
		elm$core$List$cons,
		elm$html$Html$Attributes$class('table'),
		A2(
			elm$core$List$filterMap,
			elm$core$Basics$identity,
			A2(elm$core$List$map, rundis$elm_bootstrap$Bootstrap$Table$tableClass, options)));
};
var rundis$elm_bootstrap$Bootstrap$Table$table = function (rec) {
	var isInversed = A2(
		elm$core$List$any,
		function (opt) {
			return _Utils_eq(opt, rundis$elm_bootstrap$Bootstrap$Table$Inversed);
		},
		rec.options);
	var classOptions = A2(
		elm$core$List$filter,
		function (opt) {
			return !rundis$elm_bootstrap$Bootstrap$Table$isResponsive(opt);
		},
		rec.options);
	return A2(
		rundis$elm_bootstrap$Bootstrap$Table$maybeWrapResponsive,
		rec.options,
		A2(
			elm$html$Html$table,
			rundis$elm_bootstrap$Bootstrap$Table$tableAttributes(classOptions),
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Table$renderTHead(
					A2(rundis$elm_bootstrap$Bootstrap$Table$maybeMapInversedTHead, isInversed, rec.thead)),
					rundis$elm_bootstrap$Bootstrap$Table$renderTBody(
					A2(rundis$elm_bootstrap$Bootstrap$Table$maybeMapInversedTBody, isInversed, rec.tbody))
				])));
};
var rundis$elm_bootstrap$Bootstrap$Table$tbody = F2(
	function (attributes, rows) {
		return rundis$elm_bootstrap$Bootstrap$Table$TBody(
			{attributes: attributes, rows: rows});
	});
var rundis$elm_bootstrap$Bootstrap$Table$th = F2(
	function (options, children) {
		return rundis$elm_bootstrap$Bootstrap$Table$Th(
			{children: children, options: options});
	});
var rundis$elm_bootstrap$Bootstrap$Table$thead = F2(
	function (options, rows) {
		return rundis$elm_bootstrap$Bootstrap$Table$THead(
			{options: options, rows: rows});
	});
var author$project$Routes$DeletedFiles$viewDeletedList = F2(
	function (model, entries) {
		var filteredEntries = A2(author$project$Routes$DeletedFiles$filterEntries, model.filter, entries);
		return rundis$elm_bootstrap$Bootstrap$Table$table(
			{
				options: _List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Table$hover,
						rundis$elm_bootstrap$Bootstrap$Table$attr(
						elm$html$Html$Attributes$class('borderless-table'))
					]),
				tbody: A2(
					rundis$elm_bootstrap$Bootstrap$Table$tbody,
					_List_Nil,
					A2(
						elm$core$List$map,
						author$project$Routes$DeletedFiles$viewDeletedEntry(model),
						filteredEntries)),
				thead: A2(
					rundis$elm_bootstrap$Bootstrap$Table$thead,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$tr,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '5%'))
										]),
									_List_Nil),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '55%'))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('Name')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '20%'))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('Deleted at')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '15%'))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('Size')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '5%'))
										]),
									_List_Nil)
								]))
						]))
			});
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col12);
var author$project$Routes$DeletedFiles$maybeViewDeletedList = F2(
	function (model, entries) {
		return (elm$core$List$length(entries) > 0) ? A2(author$project$Routes$DeletedFiles$viewDeletedList, model, entries) : A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsCenter)
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted')
								]),
							(!elm$core$String$length(model.filter)) ? _List_fromArray(
								[
									elm$html$Html$text(' The '),
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('fas fa-md fa-trash-alt')
										]),
									_List_Nil),
									elm$html$Html$text(' is empty. If you delete something, it will appear here.')
								]) : _List_fromArray(
								[
									elm$html$Html$text(' Search did not find anything. Remove the query to go back. ')
								]))
						]))
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Col$lg1 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$LG, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col1);
var rundis$elm_bootstrap$Bootstrap$Grid$Col$lg10 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$LG, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col10);
var author$project$Routes$DeletedFiles$viewDeletedContainer = F2(
	function (model, entries) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg1,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg10, rundis$elm_bootstrap$Bootstrap$Grid$Col$md12]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted text-center')
								]),
							_List_fromArray(
								[
									elm$html$Html$text('Deleted files')
								])),
							A2(elm$html$Html$br, _List_Nil, _List_Nil),
							A2(author$project$Util$viewAlert, author$project$Routes$DeletedFiles$AlertMsg, model.alert),
							A2(author$project$Routes$DeletedFiles$maybeViewDeletedList, model, entries),
							A2(elm$html$Html$br, _List_Nil, _List_Nil)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg1,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil)
				]));
	});
var author$project$Routes$DeletedFiles$SearchInput = function (a) {
	return {$: 'SearchInput', a: a};
};
var author$project$Routes$DeletedFiles$viewSearchBox = function (model) {
	return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$view(
		A2(
			rundis$elm_bootstrap$Bootstrap$Form$InputGroup$attrs,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('stylish-input-group input-group')
				]),
			A2(
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$successors,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Form$InputGroup$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('input-group-addon')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$button,
								_List_Nil,
								_List_fromArray(
									[
										A2(
										elm$html$Html$span,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('fas fa-search fa-xs input-group-addon')
											]),
										_List_Nil)
									]))
							]))
					]),
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$config(
					rundis$elm_bootstrap$Bootstrap$Form$InputGroup$text(
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Search'),
								rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onInput(author$project$Routes$DeletedFiles$SearchInput),
										elm$html$Html$Attributes$value(model.filter)
									]))
							]))))));
};
var author$project$Routes$DeletedFiles$view = function (model) {
	var _n0 = model.state;
	switch (_n0.$) {
		case 'Loading':
			return elm$html$Html$text('Still loading');
		case 'Failure':
			var err = _n0.a;
			return elm$html$Html$text('Failed to load log: ' + err);
		default:
			var entries = _n0.a;
			return A2(
				rundis$elm_bootstrap$Bootstrap$Grid$row,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$col,
						_List_fromArray(
							[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg12]),
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$id('main-header-row')
											]))
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl9]),
										_List_Nil),
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl3]),
										_List_fromArray(
											[
												A2(elm$html$Html$Lazy$lazy, author$project$Routes$DeletedFiles$viewSearchBox, model)
											]))
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$id('main-content-row')
											]))
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl10]),
										_List_fromArray(
											[
												A2(author$project$Routes$DeletedFiles$viewDeletedContainer, model, entries)
											]))
									]))
							]))
					]));
	}
};
var author$project$Routes$Diff$BackClicked = {$: 'BackClicked'};
var author$project$Commands$diffChangeCount = function (diff) {
	return (((((elm$core$List$length(diff.added) + elm$core$List$length(diff.removed)) + elm$core$List$length(diff.ignored)) + elm$core$List$length(diff.missing)) + elm$core$List$length(diff.moved)) + elm$core$List$length(diff.merged)) + elm$core$List$length(diff.conflict);
};
var elm$html$Html$h5 = _VirtualDom_node('h5');
var author$project$Routes$Diff$viewHeading = F2(
	function (className, message) {
		return A2(
			elm$html$Html$h5,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class(className)
				]),
			_List_fromArray(
				[
					elm$html$Html$text(message)
				]));
	});
var author$project$Routes$Diff$viewLine = function (line) {
	return A2(
		elm$html$Html$span,
		_List_Nil,
		_List_fromArray(
			[
				elm$html$Html$text(line),
				A2(elm$html$Html$br, _List_Nil, _List_Nil)
			]));
};
var author$project$Routes$Diff$viewPairs = F2(
	function (entries, header) {
		return (elm$core$List$length(entries) > 0) ? A2(
			elm$html$Html$span,
			_List_Nil,
			_List_fromArray(
				[
					header,
					A2(
					elm$html$Html$span,
					_List_Nil,
					A2(
						elm$core$List$map,
						function (p) {
							return author$project$Routes$Diff$viewLine(' ' + (p.src.path + (' ↔ ' + p.dst.path)));
						},
						entries)),
					A2(elm$html$Html$br, _List_Nil, _List_Nil)
				])) : elm$html$Html$text('');
	});
var author$project$Routes$Diff$viewSingle = F2(
	function (entries, header) {
		return (elm$core$List$length(entries) > 0) ? A2(
			elm$html$Html$span,
			_List_Nil,
			_List_fromArray(
				[
					header,
					A2(
					elm$html$Html$span,
					_List_Nil,
					A2(
						elm$core$List$map,
						function (e) {
							return author$project$Routes$Diff$viewLine(' ' + e.path);
						},
						entries)),
					A2(elm$html$Html$br, _List_Nil, _List_Nil)
				])) : elm$html$Html$text('');
	});
var author$project$Routes$Diff$viewDiff = F2(
	function (model, diff) {
		var nChanges = author$project$Commands$diffChangeCount(diff);
		if (!nChanges) {
			return elm$html$Html$text('There are no differences!');
		} else {
			var n = nChanges;
			return A2(
				elm$html$Html$div,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						author$project$Routes$Diff$viewSingle,
						diff.added,
						A2(author$project$Routes$Diff$viewHeading, 'text-success', 'Added')),
						A2(
						author$project$Routes$Diff$viewSingle,
						diff.removed,
						A2(author$project$Routes$Diff$viewHeading, 'text-warning', 'Removed')),
						A2(
						author$project$Routes$Diff$viewSingle,
						diff.ignored,
						A2(author$project$Routes$Diff$viewHeading, 'text-muted', 'Ignored')),
						A2(
						author$project$Routes$Diff$viewSingle,
						diff.missing,
						A2(author$project$Routes$Diff$viewHeading, 'text-secondary', 'Missing')),
						A2(
						author$project$Routes$Diff$viewPairs,
						diff.moved,
						A2(author$project$Routes$Diff$viewHeading, 'text-primary', 'Moved')),
						A2(
						author$project$Routes$Diff$viewPairs,
						diff.merged,
						A2(author$project$Routes$Diff$viewHeading, 'text-info', 'Merged')),
						A2(
						author$project$Routes$Diff$viewPairs,
						diff.conflict,
						A2(author$project$Routes$Diff$viewHeading, 'text-danger', 'Conflicts')),
						A2(elm$html$Html$br, _List_Nil, _List_Nil),
						A2(elm$html$Html$br, _List_Nil, _List_Nil),
						elm$html$Html$text(
						elm$core$String$fromInt(n) + ' changes in total')
					]));
		}
	});
var author$project$Routes$Diff$viewDiffContainer = F2(
	function (model, result) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg8, rundis$elm_bootstrap$Bootstrap$Grid$Col$md12]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-center')
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('text-muted')
										]),
									_List_fromArray(
										[
											elm$html$Html$text('Difference to »')
										])),
									elm$html$Html$text(
									author$project$Routes$Diff$nameFromUrl(model.url)),
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('text-muted')
										]),
									_List_fromArray(
										[
											elm$html$Html$text('«')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Button$button,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Button$roleLink,
											rundis$elm_bootstrap$Bootstrap$Button$attrs(
											_List_fromArray(
												[
													elm$html$Html$Events$onClick(author$project$Routes$Diff$BackClicked)
												]))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('font-weight-light')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('(go back)')
												]))
										]))
								])),
							A2(elm$html$Html$br, _List_Nil, _List_Nil),
							function () {
							if (result.$ === 'Ok') {
								var diff = result.a;
								return A2(author$project$Routes$Diff$viewDiff, model, diff);
							} else {
								var err = result.a;
								return elm$html$Html$text(err);
							}
						}(),
							A2(elm$html$Html$br, _List_Nil, _List_Nil)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil)
				]));
	});
var author$project$Routes$Diff$view = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Loading') {
		return elm$html$Html$text('Still loading');
	} else {
		var result = _n0.a;
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg12]),
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$row,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
									_List_fromArray(
										[
											elm$html$Html$Attributes$id('main-content-row')
										]))
								]),
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Grid$col,
									_List_fromArray(
										[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl10]),
									_List_fromArray(
										[
											A2(author$project$Routes$Diff$viewDiffContainer, model, result)
										]))
								]))
						]))
				]));
	}
};
var author$project$Modals$Mkdir$ModalShow = {$: 'ModalShow'};
var author$project$Modals$Mkdir$show = author$project$Modals$Mkdir$ModalShow;
var author$project$Modals$Remove$ModalShow = function (a) {
	return {$: 'ModalShow', a: a};
};
var author$project$Modals$Remove$show = function (paths) {
	return author$project$Modals$Remove$ModalShow(paths);
};
var author$project$Modals$Share$ModalShow = function (a) {
	return {$: 'ModalShow', a: a};
};
var author$project$Modals$Share$show = function (paths) {
	return author$project$Modals$Share$ModalShow(paths);
};
var author$project$Modals$Upload$UploadSelectedFiles = F2(
	function (a, b) {
		return {$: 'UploadSelectedFiles', a: a, b: b};
	});
var elm$file$File$decoder = _File_decoder;
var author$project$Modals$Upload$filesDecoder = A2(
	elm$json$Json$Decode$at,
	_List_fromArray(
		['target', 'files']),
	elm$json$Json$Decode$list(elm$file$File$decoder));
var elm$html$Html$label = _VirtualDom_node('label');
var elm$html$Html$Attributes$multiple = elm$html$Html$Attributes$boolProperty('multiple');
var author$project$Modals$Upload$buildButton = F4(
	function (model, currIsFile, currRoot, toMsg) {
		return A2(
			elm$html$Html$label,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('btn btn-file btn-link btn-default text-left'),
					elm$html$Html$Attributes$id('action-btn'),
					currIsFile ? elm$html$Html$Attributes$class('disabled') : elm$html$Html$Attributes$class('btn-default')
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fas fa-plus')
						]),
					_List_Nil),
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('d-lg-inline d-none')
						]),
					_List_fromArray(
						[
							elm$html$Html$text('\u00a0\u00a0Upload')
						])),
					A2(
					elm$html$Html$input,
					_List_fromArray(
						[
							elm$html$Html$Attributes$type_('file'),
							elm$html$Html$Attributes$multiple(true),
							A2(
							elm$html$Html$Events$on,
							'change',
							A2(
								elm$json$Json$Decode$map,
								toMsg,
								A2(
									elm$json$Json$Decode$map,
									author$project$Modals$Upload$UploadSelectedFiles(currRoot),
									author$project$Modals$Upload$filesDecoder))),
							A2(elm$html$Html$Attributes$style, 'display', 'none'),
							elm$html$Html$Attributes$disabled(currIsFile)
						]),
					_List_Nil)
				]));
	});
var author$project$Modals$Upload$clampText = F2(
	function (text, length) {
		return (_Utils_cmp(
			elm$core$String$length(text),
			length) < 1) ? text : (A3(elm$core$String$slice, 0, length, text) + '…');
	});
var author$project$Modals$Upload$viewAlert = F3(
	function (alert, path, isSuccess) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Alert$view,
			alert,
			A2(
				rundis$elm_bootstrap$Bootstrap$Alert$children,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$row,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs10]),
								_List_fromArray(
									[
										A2(
										elm$html$Html$span,
										_List_fromArray(
											[
												isSuccess ? elm$html$Html$Attributes$class('fas fa-xs fa-check') : elm$html$Html$Attributes$class('fas fa-xs fa-exclamation-circle')
											]),
										_List_Nil),
										elm$html$Html$text(
										' ' + A2(author$project$Modals$Upload$clampText, path, 15))
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Col$xs2,
										rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsRight)
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Button$button,
										_List_fromArray(
											[
												rundis$elm_bootstrap$Bootstrap$Button$roleLink,
												rundis$elm_bootstrap$Bootstrap$Button$attrs(
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('notification-close-btn'),
														elm$html$Html$Events$onClick(
														A2(author$project$Modals$Upload$AlertMsg, path, rundis$elm_bootstrap$Bootstrap$Alert$closed))
													]))
											]),
										_List_fromArray(
											[
												A2(
												elm$html$Html$span,
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('fas fa-xs fa-times')
													]),
												_List_Nil)
											]))
									]))
							]))
					]),
				(isSuccess ? rundis$elm_bootstrap$Bootstrap$Alert$success : rundis$elm_bootstrap$Bootstrap$Alert$danger)(
					A2(
						rundis$elm_bootstrap$Bootstrap$Alert$dismissableWithAnimation,
						author$project$Modals$Upload$AlertMsg(path),
						rundis$elm_bootstrap$Bootstrap$Alert$config))));
	});
var author$project$Modals$Upload$UploadCancel = function (a) {
	return {$: 'UploadCancel', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$md10 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$MD, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col10);
var rundis$elm_bootstrap$Bootstrap$Grid$Col$md2 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$MD, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col2);
var rundis$elm_bootstrap$Bootstrap$Progress$Attrs = function (a) {
	return {$: 'Attrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Progress$attrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Progress$Attrs(attrs_);
};
var rundis$elm_bootstrap$Bootstrap$Progress$Label = function (a) {
	return {$: 'Label', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Progress$customLabel = function (children) {
	return rundis$elm_bootstrap$Bootstrap$Progress$Label(children);
};
var rundis$elm_bootstrap$Bootstrap$Progress$Options = function (a) {
	return {$: 'Options', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Progress$applyOption = F2(
	function (modifier, _n0) {
		var options = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Progress$Options(
			function () {
				switch (modifier.$) {
					case 'Value':
						var value_ = modifier.a;
						return _Utils_update(
							options,
							{value: value_});
					case 'Height':
						var height_ = modifier.a;
						return _Utils_update(
							options,
							{height: height_});
					case 'Label':
						var label_ = modifier.a;
						return _Utils_update(
							options,
							{label: label_});
					case 'Roled':
						var role_ = modifier.a;
						return _Utils_update(
							options,
							{role: role_});
					case 'Striped':
						var striped_ = modifier.a;
						return _Utils_update(
							options,
							{striped: striped_});
					case 'Animated':
						var animated_ = modifier.a;
						return _Utils_update(
							options,
							{animated: animated_});
					case 'Attrs':
						var attrs_ = modifier.a;
						return _Utils_update(
							options,
							{attributes: attrs_});
					default:
						var attrs_ = modifier.a;
						return _Utils_update(
							options,
							{wrapperAttributes: attrs_});
				}
			}());
	});
var rundis$elm_bootstrap$Bootstrap$Progress$defaultOptions = rundis$elm_bootstrap$Bootstrap$Progress$Options(
	{animated: false, attributes: _List_Nil, height: elm$core$Maybe$Nothing, label: _List_Nil, role: elm$core$Maybe$Nothing, striped: false, value: 0, wrapperAttributes: _List_Nil});
var rundis$elm_bootstrap$Bootstrap$Progress$roleClass = function (role) {
	return elm$html$Html$Attributes$class(
		function () {
			switch (role.$) {
				case 'Success':
					return 'bg-success';
				case 'Info':
					return 'bg-info';
				case 'Warning':
					return 'bg-warning';
				default:
					return 'bg-danger';
			}
		}());
};
var rundis$elm_bootstrap$Bootstrap$Progress$toAttributes = function (_n0) {
	var options = _n0.a;
	return elm$core$List$concat(
		_List_fromArray(
			[
				_List_fromArray(
				[
					A2(elm$html$Html$Attributes$attribute, 'role', 'progressbar'),
					A2(
					elm$html$Html$Attributes$attribute,
					'aria-value-now',
					elm$core$String$fromFloat(options.value)),
					A2(elm$html$Html$Attributes$attribute, 'aria-valuemin', '0'),
					A2(elm$html$Html$Attributes$attribute, 'aria-valuemax', '100'),
					A2(
					elm$html$Html$Attributes$style,
					'width',
					elm$core$String$fromFloat(options.value) + '%'),
					elm$html$Html$Attributes$classList(
					_List_fromArray(
						[
							_Utils_Tuple2('progress-bar', true),
							_Utils_Tuple2('progress-bar-striped', options.striped || options.animated),
							_Utils_Tuple2('progress-bar-animated', options.animated)
						]))
				]),
				function () {
				var _n1 = options.height;
				if (_n1.$ === 'Just') {
					var height_ = _n1.a;
					return _List_fromArray(
						[
							A2(
							elm$html$Html$Attributes$style,
							'height',
							elm$core$String$fromInt(height_) + 'px')
						]);
				} else {
					return _List_Nil;
				}
			}(),
				function () {
				var _n2 = options.role;
				if (_n2.$ === 'Just') {
					var role_ = _n2.a;
					return _List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Progress$roleClass(role_)
						]);
				} else {
					return _List_Nil;
				}
			}(),
				options.attributes
			]));
};
var rundis$elm_bootstrap$Bootstrap$Progress$renderBar = function (modifiers) {
	var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Progress$applyOption, rundis$elm_bootstrap$Bootstrap$Progress$defaultOptions, modifiers);
	var opts = options.a;
	return A2(
		elm$html$Html$div,
		rundis$elm_bootstrap$Bootstrap$Progress$toAttributes(options),
		opts.label);
};
var rundis$elm_bootstrap$Bootstrap$Progress$progress = function (modifiers) {
	var _n0 = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Progress$applyOption, rundis$elm_bootstrap$Bootstrap$Progress$defaultOptions, modifiers);
	var options = _n0.a;
	return A2(
		elm$html$Html$div,
		A2(
			elm$core$List$cons,
			elm$html$Html$Attributes$class('progress'),
			options.wrapperAttributes),
		_List_fromArray(
			[
				rundis$elm_bootstrap$Bootstrap$Progress$renderBar(modifiers)
			]));
};
var rundis$elm_bootstrap$Bootstrap$Progress$Value = function (a) {
	return {$: 'Value', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Progress$value = function (val) {
	return rundis$elm_bootstrap$Bootstrap$Progress$Value(val);
};
var rundis$elm_bootstrap$Bootstrap$Progress$WrapperAttrs = function (a) {
	return {$: 'WrapperAttrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Progress$wrapperAttrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Progress$WrapperAttrs(attrs_);
};
var author$project$Modals$Upload$viewProgressIndicator = F2(
	function (path, fraction) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$md10]),
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Progress$progress(
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Progress$value(100 * fraction),
									rundis$elm_bootstrap$Bootstrap$Progress$customLabel(
									_List_fromArray(
										[
											elm$html$Html$text(
											A2(author$project$Modals$Upload$clampText, path, 25))
										])),
									rundis$elm_bootstrap$Bootstrap$Progress$attrs(
									_List_fromArray(
										[
											A2(elm$html$Html$Attributes$style, 'height', '25px')
										])),
									rundis$elm_bootstrap$Bootstrap$Progress$wrapperAttrs(
									_List_fromArray(
										[
											A2(elm$html$Html$Attributes$style, 'height', '25px')
										]))
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$md2]),
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Button$button,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Button$roleLink,
									rundis$elm_bootstrap$Bootstrap$Button$attrs(
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('progress-cancel'),
											elm$html$Html$Events$onClick(
											author$project$Modals$Upload$UploadCancel(path))
										]))
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('fas fa-xs fa-times')
										]),
									_List_Nil)
								]))
						]))
				]));
	});
var author$project$Modals$Upload$viewUploadState = function (model) {
	return A2(
		elm$html$Html$div,
		_List_Nil,
		_List_fromArray(
			[
				A2(elm$html$Html$br, _List_Nil, _List_Nil),
				A2(elm$html$Html$br, _List_Nil, _List_Nil),
				A2(
				elm$html$Html$ul,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('notification-list list-group')
					]),
				_Utils_ap(
					A2(
						elm$core$List$map,
						function (a) {
							return A3(author$project$Modals$Upload$viewAlert, a.alert, a.path, true);
						},
						model.success),
					_Utils_ap(
						A2(
							elm$core$List$map,
							function (a) {
								return A3(author$project$Modals$Upload$viewAlert, a.alert, a.path, false);
							},
							model.failed),
						A2(
							elm$core$List$map,
							function (p) {
								return A2(author$project$Modals$Upload$viewProgressIndicator, p.a, p.b);
							},
							elm$core$Dict$toList(model.uploads)))))
			]));
};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Block = {$: 'Block'};
var rundis$elm_bootstrap$Bootstrap$Button$block = rundis$elm_bootstrap$Bootstrap$Internal$Button$Block;
var rundis$elm_bootstrap$Bootstrap$General$Internal$SM = {$: 'SM'};
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Size = function (a) {
	return {$: 'Size', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Button$small = rundis$elm_bootstrap$Bootstrap$Internal$Button$Size(rundis$elm_bootstrap$Bootstrap$General$Internal$SM);
var author$project$Routes$Ls$buildActionButton = F4(
	function (msg, iconName, labelText, isDisabled) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Button$button,
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Button$block,
					rundis$elm_bootstrap$Bootstrap$Button$small,
					rundis$elm_bootstrap$Bootstrap$Button$roleLink,
					rundis$elm_bootstrap$Bootstrap$Button$attrs(
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('text-left'),
							elm$html$Html$Attributes$disabled(isDisabled),
							elm$html$Html$Events$onClick(msg)
						]))
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fas fa-lg'),
							elm$html$Html$Attributes$class(iconName)
						]),
					_List_Nil),
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('d-lg-inline d-none')
						]),
					_List_fromArray(
						[
							elm$html$Html$text(' ' + labelText)
						]))
				]));
	});
var author$project$Routes$Ls$currIsFile = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		return !actualModel.self.isDir;
	} else {
		return false;
	}
};
var author$project$Routes$Ls$currRoot = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		return elm$core$Maybe$Just(actualModel.self.path);
	} else {
		return elm$core$Maybe$Nothing;
	}
};
var elm$core$Dict$member = F2(
	function (key, dict) {
		var _n0 = A2(elm$core$Dict$get, key, dict);
		if (_n0.$ === 'Just') {
			return true;
		} else {
			return false;
		}
	});
var elm$core$Set$member = F2(
	function (key, _n0) {
		var dict = _n0.a;
		return A2(elm$core$Dict$member, key, dict);
	});
var author$project$Routes$Ls$currSelectedSize = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		var entryToSizeIfSelected = function (e) {
			return A2(elm$core$Set$member, e.path, actualModel.checked) ? e.size : 0;
		};
		return A3(
			elm$core$List$foldl,
			elm$core$Basics$add,
			0,
			A2(elm$core$List$map, entryToSizeIfSelected, actualModel.entries));
	} else {
		return 0;
	}
};
var author$project$Routes$Ls$currTotalSize = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		return actualModel.self.size;
	} else {
		return 0;
	}
};
var elm$html$Html$p = _VirtualDom_node('p');
var author$project$Routes$Ls$labelSelectedItems = F2(
	function (model, num) {
		if (author$project$Routes$Ls$currIsFile(model)) {
			return elm$html$Html$text('');
		} else {
			switch (num) {
				case 0:
					return A2(
						elm$html$Html$p,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text(' Nothing selected'),
								A2(elm$html$Html$br, _List_Nil, _List_Nil),
								elm$html$Html$text(
								basti1302$elm_human_readable_filesize$Filesize$format(
									author$project$Routes$Ls$currTotalSize(model)) + ' in total')
							]));
				case 1:
					return A2(
						elm$html$Html$p,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text(' 1 item'),
								A2(elm$html$Html$br, _List_Nil, _List_Nil),
								elm$html$Html$text(
								basti1302$elm_human_readable_filesize$Filesize$format(
									author$project$Routes$Ls$currSelectedSize(model)))
							]));
				default:
					var n = num;
					return A2(
						elm$html$Html$p,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text(
								' ' + (elm$core$String$fromInt(n) + ' items')),
								A2(elm$html$Html$br, _List_Nil, _List_Nil),
								elm$html$Html$text(
								basti1302$elm_human_readable_filesize$Filesize$format(
									author$project$Routes$Ls$currSelectedSize(model)))
							]));
			}
		}
	});
var elm$core$Dict$filter = F2(
	function (isGood, dict) {
		return A3(
			elm$core$Dict$foldl,
			F3(
				function (k, v, d) {
					return A2(isGood, k, v) ? A3(elm$core$Dict$insert, k, v, d) : d;
				}),
			elm$core$Dict$empty,
			dict);
	});
var elm$core$Set$filter = F2(
	function (isGood, _n0) {
		var dict = _n0.a;
		return elm$core$Set$Set_elm_builtin(
			A2(
				elm$core$Dict$filter,
				F2(
					function (key, _n1) {
						return isGood(key);
					}),
				dict));
	});
var author$project$Routes$Ls$nSelectedItems = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		return elm$core$Set$size(
			A2(
				elm$core$Set$filter,
				function (e) {
					return !elm$core$String$isEmpty(e);
				},
				actualModel.checked));
	} else {
		return 0;
	}
};
var author$project$Routes$Ls$selectedPaths = function (model) {
	var _n0 = model.state;
	if (_n0.$ === 'Success') {
		var actualModel = _n0.a;
		return elm$core$Set$toList(
			A2(
				elm$core$Set$filter,
				function (e) {
					return !elm$core$String$isEmpty(e);
				},
				actualModel.checked));
	} else {
		return _List_Nil;
	}
};
var elm$url$Url$Builder$absolute = F2(
	function (pathSegments, parameters) {
		return '/' + (A2(elm$core$String$join, '/', pathSegments) + elm$url$Url$Builder$toQuery(parameters));
	});
var author$project$Routes$Ls$buildDownloadUrl = function (model) {
	return A2(
		elm$url$Url$Builder$absolute,
		A2(
			elm$core$List$cons,
			'get',
			author$project$Util$splitPath(
				author$project$Util$urlToPath(model.url))),
		A2(
			elm$core$List$cons,
			A2(elm$url$Url$Builder$string, 'direct', 'yes'),
			(author$project$Routes$Ls$nSelectedItems(model) > 0) ? A2(
				elm$core$List$map,
				elm$url$Url$Builder$string('include'),
				author$project$Routes$Ls$selectedPaths(model)) : _List_Nil));
};
var rundis$elm_bootstrap$Bootstrap$Button$linkButton = F2(
	function (options, children) {
		return A2(
			elm$html$Html$a,
			A2(
				elm$core$List$cons,
				A2(elm$html$Html$Attributes$attribute, 'role', 'button'),
				rundis$elm_bootstrap$Bootstrap$Internal$Button$buttonAttributes(options)),
			children);
	});
var author$project$Routes$Ls$viewSidebarDownloadButton = function (model) {
	var nSelected = author$project$Routes$Ls$nSelectedItems(model);
	var disabledClass = (author$project$Routes$Ls$currIsFile(model) || (!A2(elm$core$List$member, 'fs.download', model.rights))) ? elm$html$Html$Attributes$class('disabled') : elm$html$Html$Attributes$class('btn-default');
	return A2(
		rundis$elm_bootstrap$Bootstrap$Button$linkButton,
		_List_fromArray(
			[
				rundis$elm_bootstrap$Bootstrap$Button$block,
				rundis$elm_bootstrap$Bootstrap$Button$attrs(
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('text-left btn-link download-btn'),
						disabledClass,
						elm$html$Html$Attributes$href(
						author$project$Routes$Ls$buildDownloadUrl(model))
					]))
			]),
		_List_fromArray(
			[
				A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fas fa-lg fa-file-download')
					]),
				_List_Nil),
				A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$id('action-btn'),
						elm$html$Html$Attributes$class('d-none d-lg-inline')
					]),
				_List_fromArray(
					[
						(nSelected > 0) ? elm$html$Html$text(' Download selected ') : elm$html$Html$text(' Download all')
					]))
			]));
};
var elm$virtual_dom$VirtualDom$map = _VirtualDom_map;
var elm$html$Html$map = elm$virtual_dom$VirtualDom$map;
var author$project$Routes$Ls$viewActionList = function (model) {
	var root = A2(
		elm$core$Maybe$withDefault,
		'/',
		author$project$Routes$Ls$currRoot(model));
	var nSelected = author$project$Routes$Ls$nSelectedItems(model);
	return A2(
		elm$html$Html$div,
		_List_Nil,
		_List_fromArray(
			[
				A2(
				elm$html$Html$div,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('d-flex flex-lg-column flex-row')
					]),
				_List_fromArray(
					[
						A2(
						elm$html$Html$p,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('text-muted'),
								elm$html$Html$Attributes$id('select-label')
							]),
						_List_fromArray(
							[
								A2(author$project$Routes$Ls$labelSelectedItems, model, nSelected)
							])),
						A2(
						elm$html$Html$div,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('d-flex flex-column')
							]),
						_List_fromArray(
							[
								A4(
								author$project$Modals$Upload$buildButton,
								model.uploadState,
								author$project$Routes$Ls$currIsFile(model) || (!A2(elm$core$List$member, 'fs.download', model.rights)),
								root,
								author$project$Routes$Ls$UploadMsg),
								author$project$Routes$Ls$viewSidebarDownloadButton(model)
							])),
						A2(
						elm$html$Html$div,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('d-flex flex-column')
							]),
						_List_fromArray(
							[
								A4(
								author$project$Routes$Ls$buildActionButton,
								author$project$Routes$Ls$ShareMsg(
									author$project$Modals$Share$show(
										author$project$Routes$Ls$selectedPaths(model))),
								'fa-share-alt',
								'Share',
								!nSelected),
								A4(
								author$project$Routes$Ls$buildActionButton,
								author$project$Routes$Ls$MkdirMsg(author$project$Modals$Mkdir$show),
								'fa-edit',
								'New Folder',
								author$project$Routes$Ls$currIsFile(model) || (!A2(elm$core$List$member, 'fs.edit', model.rights)))
							])),
						A2(
						elm$html$Html$div,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('d-flex flex-column')
							]),
						_List_fromArray(
							[
								A4(
								author$project$Routes$Ls$buildActionButton,
								author$project$Routes$Ls$RemoveMsg(
									author$project$Modals$Remove$show(
										author$project$Routes$Ls$selectedPaths(model))),
								'fa-trash',
								'Delete',
								author$project$Routes$Ls$currIsFile(model) || ((!nSelected) || (!A2(elm$core$List$member, 'fs.edit', model.rights))))
							]))
					])),
				A2(
				elm$html$Html$div,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						elm$html$Html$map,
						author$project$Routes$Ls$UploadMsg,
						author$project$Modals$Upload$viewUploadState(model.uploadState))
					]))
			]));
};
var rundis$elm_bootstrap$Bootstrap$Breadcrumb$Item = F2(
	function (a, b) {
		return {$: 'Item', a: a, b: b};
	});
var rundis$elm_bootstrap$Bootstrap$Breadcrumb$item = F2(
	function (attributes, children) {
		return A2(rundis$elm_bootstrap$Bootstrap$Breadcrumb$Item, attributes, children);
	});
var author$project$Routes$Ls$buildBreadcrumbs = F2(
	function (names, previous) {
		var displayName = function (n) {
			return (elm$core$String$length(n) <= 0) ? 'Home' : n;
		};
		if (!names.b) {
			return _List_Nil;
		} else {
			if (!names.b.b) {
				var name = names.a;
				return _List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Breadcrumb$item,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text(
								displayName(name))
							]))
					]);
			} else {
				var name = names.a;
				var rest = names.b;
				return A2(
					elm$core$List$cons,
					A2(
						rundis$elm_bootstrap$Bootstrap$Breadcrumb$item,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								elm$html$Html$a,
								_List_fromArray(
									[
										elm$html$Html$Attributes$href(
										'/view/' + A2(
											elm$core$String$join,
											'/',
											A2(elm$core$List$cons, name, previous)))
									]),
								_List_fromArray(
									[
										elm$html$Html$text(
										displayName(name))
									]))
							])),
					A2(
						author$project$Routes$Ls$buildBreadcrumbs,
						rest,
						_Utils_ap(
							previous,
							_List_fromArray(
								[name]))));
			}
		}
	});
var elm$html$Html$nav = _VirtualDom_node('nav');
var elm$html$Html$ol = _VirtualDom_node('ol');
var rundis$elm_bootstrap$Bootstrap$Breadcrumb$toListItems = function (items) {
	if (!items.b) {
		return _List_Nil;
	} else {
		if (!items.b.b) {
			var _n1 = items.a;
			var attributes = _n1.a;
			var children = _n1.b;
			return _List_fromArray(
				[
					A2(
					elm$html$Html$li,
					_Utils_ap(
						attributes,
						_List_fromArray(
							[
								A2(elm$html$Html$Attributes$attribute, 'aria-current', 'page'),
								elm$html$Html$Attributes$class('breadcrumb-item active')
							])),
					children)
				]);
		} else {
			var _n2 = items.a;
			var attributes = _n2.a;
			var children = _n2.b;
			var rest = items.b;
			return _Utils_ap(
				_List_fromArray(
					[
						A2(
						elm$html$Html$li,
						_Utils_ap(
							attributes,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('breadcrumb-item')
								])),
						children)
					]),
				rundis$elm_bootstrap$Bootstrap$Breadcrumb$toListItems(rest));
		}
	}
};
var rundis$elm_bootstrap$Bootstrap$Breadcrumb$container = function (items) {
	if (!items.b) {
		return elm$html$Html$text('');
	} else {
		return A2(
			elm$html$Html$nav,
			_List_fromArray(
				[
					A2(elm$html$Html$Attributes$attribute, 'aria-label', 'breadcrumb'),
					A2(elm$html$Html$Attributes$attribute, 'role', 'navigation')
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$ol,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('breadcrumb')
						]),
					rundis$elm_bootstrap$Bootstrap$Breadcrumb$toListItems(items))
				]));
	}
};
var author$project$Routes$Ls$viewBreadcrumbs = function (model) {
	return A2(
		elm$html$Html$div,
		_List_fromArray(
			[
				elm$html$Html$Attributes$id('breadcrumbs-box')
			]),
		_List_fromArray(
			[
				rundis$elm_bootstrap$Bootstrap$Breadcrumb$container(
				A2(
					author$project$Routes$Ls$buildBreadcrumbs,
					A2(
						elm$core$List$cons,
						'',
						author$project$Util$splitPath(
							author$project$Util$urlToPath(model.url))),
					_List_Nil))
			]));
};
var author$project$Routes$Ls$CheckboxTickAll = function (a) {
	return {$: 'CheckboxTickAll', a: a};
};
var author$project$Routes$Ls$ModTime = {$: 'ModTime'};
var author$project$Routes$Ls$Name = {$: 'Name'};
var author$project$Routes$Ls$Pin = {$: 'Pin'};
var author$project$Routes$Ls$Size = {$: 'Size'};
var author$project$Routes$Ls$SortBy = F2(
	function (a, b) {
		return {$: 'SortBy', a: a, b: b};
	});
var author$project$Routes$Ls$buildSortControl = F3(
	function (name, model, key) {
		var descClass = _Utils_eq(
			_Utils_Tuple2(author$project$Routes$Ls$Descending, key),
			model.sortState) ? 'sort-button-selected' : '';
		var ascClass = _Utils_eq(
			_Utils_Tuple2(author$project$Routes$Ls$Ascending, key),
			model.sortState) ? 'sort-button-selected' : '';
		return A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('sort-button-container text-muted')
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text(name + ' ')
						])),
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('sort-button')
						]),
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Button$linkButton,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Button$small,
									rundis$elm_bootstrap$Bootstrap$Button$attrs(
									_List_fromArray(
										[
											elm$html$Html$Events$onClick(
											A2(author$project$Routes$Ls$SortBy, author$project$Routes$Ls$Ascending, key)),
											elm$html$Html$Attributes$class('sort-button')
										]))
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('fas fa-xs fa-arrow-up'),
											elm$html$Html$Attributes$class(ascClass)
										]),
									_List_Nil)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Button$linkButton,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Button$small,
									rundis$elm_bootstrap$Bootstrap$Button$attrs(
									_List_fromArray(
										[
											elm$html$Html$Events$onClick(
											A2(author$project$Routes$Ls$SortBy, author$project$Routes$Ls$Descending, key)),
											elm$html$Html$Attributes$class('sort-button')
										]))
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('fas fa-xs fa-arrow-down'),
											elm$html$Html$Attributes$class(descClass)
										]),
									_List_Nil)
								]))
						]))
				]));
	});
var author$project$Routes$Ls$CheckboxTick = F2(
	function (a, b) {
		return {$: 'CheckboxTick', a: a, b: b};
	});
var author$project$Routes$Ls$RowClicked = function (a) {
	return {$: 'RowClicked', a: a};
};
var author$project$Modals$MoveCopy$ModalShow = function (a) {
	return {$: 'ModalShow', a: a};
};
var author$project$Modals$MoveCopy$show = function (sourcePath) {
	return author$project$Modals$MoveCopy$ModalShow(sourcePath);
};
var author$project$Modals$Rename$ModalShow = function (a) {
	return {$: 'ModalShow', a: a};
};
var author$project$Modals$Rename$show = function (currPath) {
	return author$project$Modals$Rename$ModalShow(currPath);
};
var author$project$Routes$Ls$HistoryClicked = function (a) {
	return {$: 'HistoryClicked', a: a};
};
var author$project$Routes$Ls$RemoveClicked = function (a) {
	return {$: 'RemoveClicked', a: a};
};
var author$project$Routes$Ls$mayDownload = function (model) {
	return A2(elm$core$List$member, 'fs.download', model.rights);
};
var author$project$Routes$Ls$mayEdit = function (model) {
	return A2(elm$core$List$member, 'fs.edit', model.rights);
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$DropdownItem = function (a) {
	return {$: 'DropdownItem', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$anchorItem = F2(
	function (attributes, children) {
		return rundis$elm_bootstrap$Bootstrap$Dropdown$DropdownItem(
			A2(
				elm$html$Html$a,
				_Utils_ap(
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('dropdown-item')
						]),
					attributes),
				children));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem = F2(
	function (attributes, children) {
		return rundis$elm_bootstrap$Bootstrap$Dropdown$DropdownItem(
			A2(
				elm$html$Html$button,
				_Utils_ap(
					_List_fromArray(
						[
							elm$html$Html$Attributes$type_('button'),
							elm$html$Html$Attributes$class('dropdown-item')
						]),
					attributes),
				children));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$divider = rundis$elm_bootstrap$Bootstrap$Dropdown$DropdownItem(
	A2(
		elm$html$Html$div,
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('dropdown-divider')
			]),
		_List_Nil));
var rundis$elm_bootstrap$Bootstrap$Dropdown$dropDir = function (maybeDir) {
	var toAttrs = function (dir) {
		return _List_fromArray(
			[
				elm$html$Html$Attributes$class(
				'drop' + function () {
					if (dir.$ === 'Dropleft') {
						return 'left';
					} else {
						return 'right';
					}
				}())
			]);
	};
	return A2(
		elm$core$Maybe$withDefault,
		_List_Nil,
		A2(elm$core$Maybe$map, toAttrs, maybeDir));
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$dropdownAttributes = F2(
	function (status, config) {
		return _Utils_ap(
			_List_fromArray(
				[
					elm$html$Html$Attributes$classList(
					_List_fromArray(
						[
							_Utils_Tuple2('btn-group', true),
							_Utils_Tuple2(
							'show',
							!_Utils_eq(status, rundis$elm_bootstrap$Bootstrap$Dropdown$Closed)),
							_Utils_Tuple2('dropup', config.isDropUp)
						]))
				]),
			_Utils_ap(
				rundis$elm_bootstrap$Bootstrap$Dropdown$dropDir(config.dropDirection),
				config.attributes));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$menuStyles = F2(
	function (_n0, config) {
		var status = _n0.a.status;
		var toggleSize = _n0.a.toggleSize;
		var menuSize = _n0.a.menuSize;
		var px = function (n) {
			return elm$core$String$fromFloat(n) + 'px';
		};
		var translate = F3(
			function (x, y, z) {
				return 'translate3d(' + (px(x) + (',' + (px(y) + (',' + (px(z) + ')')))));
			});
		var _default = _List_fromArray(
			[
				A2(elm$html$Html$Attributes$style, 'top', '0'),
				A2(elm$html$Html$Attributes$style, 'left', '0')
			]);
		var _n1 = _Utils_Tuple2(config.isDropUp, config.dropDirection);
		_n1$0:
		while (true) {
			if (_n1.b.$ === 'Just') {
				if (_n1.b.a.$ === 'Dropright') {
					if (_n1.a) {
						break _n1$0;
					} else {
						var _n2 = _n1.b.a;
						return _default;
					}
				} else {
					if (_n1.a) {
						break _n1$0;
					} else {
						var _n3 = _n1.b.a;
						return _Utils_ap(
							_default,
							_List_fromArray(
								[
									A2(
									elm$html$Html$Attributes$style,
									'transform',
									A3(translate, (-toggleSize.width) - menuSize.width, 0, 0))
								]));
					}
				}
			} else {
				if (_n1.a) {
					break _n1$0;
				} else {
					return _Utils_ap(
						_default,
						_List_fromArray(
							[
								A2(
								elm$html$Html$Attributes$style,
								'transform',
								A3(translate, -toggleSize.width, toggleSize.height, 0))
							]));
				}
			}
		}
		return _Utils_ap(
			_default,
			_List_fromArray(
				[
					A2(
					elm$html$Html$Attributes$style,
					'transform',
					A3(translate, -toggleSize.width, -menuSize.height, 0))
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$dropdownMenu = F3(
	function (state, config, items) {
		var status = state.a.status;
		var menuSize = state.a.menuSize;
		var wrapperStyles = _Utils_eq(status, rundis$elm_bootstrap$Bootstrap$Dropdown$Closed) ? _List_fromArray(
			[
				A2(elm$html$Html$Attributes$style, 'height', '0'),
				A2(elm$html$Html$Attributes$style, 'overflow', 'hidden'),
				A2(elm$html$Html$Attributes$style, 'position', 'relative')
			]) : _List_fromArray(
			[
				A2(elm$html$Html$Attributes$style, 'position', 'relative')
			]);
		return A2(
			elm$html$Html$div,
			wrapperStyles,
			_List_fromArray(
				[
					A2(
					elm$html$Html$div,
					_Utils_ap(
						_List_fromArray(
							[
								elm$html$Html$Attributes$classList(
								_List_fromArray(
									[
										_Utils_Tuple2('dropdown-menu', true),
										_Utils_Tuple2('dropdown-menu-right', config.hasMenuRight),
										_Utils_Tuple2('show', true)
									]))
							]),
						_Utils_ap(
							A2(rundis$elm_bootstrap$Bootstrap$Dropdown$menuStyles, state, config),
							config.menuAttrs)),
					A2(
						elm$core$List$map,
						function (_n0) {
							var x = _n0.a;
							return x;
						},
						items))
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$applyModifier = F2(
	function (option, options) {
		switch (option.$) {
			case 'AlignMenuRight':
				return _Utils_update(
					options,
					{hasMenuRight: true});
			case 'Dropup':
				return _Utils_update(
					options,
					{isDropUp: true});
			case 'Attrs':
				var attrs_ = option.a;
				return _Utils_update(
					options,
					{attributes: attrs_});
			case 'DropToDir':
				var dir = option.a;
				return _Utils_update(
					options,
					{
						dropDirection: elm$core$Maybe$Just(dir)
					});
			default:
				var attrs_ = option.a;
				return _Utils_update(
					options,
					{menuAttrs: attrs_});
		}
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$defaultOptions = {attributes: _List_Nil, dropDirection: elm$core$Maybe$Nothing, hasMenuRight: false, isDropUp: false, menuAttrs: _List_Nil};
var rundis$elm_bootstrap$Bootstrap$Dropdown$toConfig = function (options) {
	return A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$Dropdown$applyModifier, rundis$elm_bootstrap$Bootstrap$Dropdown$defaultOptions, options);
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$dropdown = F2(
	function (state, _n0) {
		var status = state.a.status;
		var toggleMsg = _n0.toggleMsg;
		var toggleButton = _n0.toggleButton;
		var items = _n0.items;
		var options = _n0.options;
		var config = rundis$elm_bootstrap$Bootstrap$Dropdown$toConfig(options);
		var _n1 = toggleButton;
		var buttonFn = _n1.a;
		return A2(
			elm$html$Html$div,
			A2(rundis$elm_bootstrap$Bootstrap$Dropdown$dropdownAttributes, status, config),
			_List_fromArray(
				[
					A2(buttonFn, toggleMsg, state),
					A3(rundis$elm_bootstrap$Bootstrap$Dropdown$dropdownMenu, state, config, items)
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$DropdownToggle = function (a) {
	return {$: 'DropdownToggle', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$Open = {$: 'Open'};
var rundis$elm_bootstrap$Bootstrap$Dropdown$nextStatus = function (status) {
	switch (status.$) {
		case 'Open':
			return rundis$elm_bootstrap$Bootstrap$Dropdown$Closed;
		case 'ListenClicks':
			return rundis$elm_bootstrap$Bootstrap$Dropdown$Closed;
		default:
			return rundis$elm_bootstrap$Bootstrap$Dropdown$Open;
	}
};
var elm$core$Tuple$pair = F2(
	function (a, b) {
		return _Utils_Tuple2(a, b);
	});
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$className = A2(
	elm$json$Json$Decode$at,
	_List_fromArray(
		['className']),
	elm$json$Json$Decode$string);
var rundis$elm_bootstrap$Bootstrap$Dropdown$isToggle = A2(
	elm$json$Json$Decode$andThen,
	function (_class) {
		return A2(elm$core$String$contains, 'dropdown-toggle', _class) ? elm$json$Json$Decode$succeed(true) : elm$json$Json$Decode$succeed(false);
	},
	rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$className);
var rundis$elm_bootstrap$Bootstrap$Dropdown$toggler = F2(
	function (path, decoder) {
		return elm$json$Json$Decode$oneOf(
			_List_fromArray(
				[
					A2(
					elm$json$Json$Decode$andThen,
					function (res) {
						return res ? A2(elm$json$Json$Decode$at, path, decoder) : elm$json$Json$Decode$fail('');
					},
					A2(elm$json$Json$Decode$at, path, rundis$elm_bootstrap$Bootstrap$Dropdown$isToggle)),
					A2(
					elm$json$Json$Decode$andThen,
					function (_n0) {
						return A2(
							rundis$elm_bootstrap$Bootstrap$Dropdown$toggler,
							_Utils_ap(
								path,
								_List_fromArray(
									['parentElement'])),
							decoder);
					},
					A2(
						elm$json$Json$Decode$at,
						_Utils_ap(
							path,
							_List_fromArray(
								['parentElement'])),
						rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$className)),
					elm$json$Json$Decode$fail('No toggler found')
				]));
	});
var elm$json$Json$Decode$float = _Json_decodeFloat;
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetHeight = A2(elm$json$Json$Decode$field, 'offsetHeight', elm$json$Json$Decode$float);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetWidth = A2(elm$json$Json$Decode$field, 'offsetWidth', elm$json$Json$Decode$float);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetLeft = A2(elm$json$Json$Decode$field, 'offsetLeft', elm$json$Json$Decode$float);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetParent = F2(
	function (x, decoder) {
		return elm$json$Json$Decode$oneOf(
			_List_fromArray(
				[
					A2(
					elm$json$Json$Decode$field,
					'offsetParent',
					elm$json$Json$Decode$null(x)),
					A2(elm$json$Json$Decode$field, 'offsetParent', decoder)
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetTop = A2(elm$json$Json$Decode$field, 'offsetTop', elm$json$Json$Decode$float);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$scrollLeft = A2(elm$json$Json$Decode$field, 'scrollLeft', elm$json$Json$Decode$float);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$scrollTop = A2(elm$json$Json$Decode$field, 'scrollTop', elm$json$Json$Decode$float);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$position = F2(
	function (x, y) {
		return A2(
			elm$json$Json$Decode$andThen,
			function (_n0) {
				var x_ = _n0.a;
				var y_ = _n0.b;
				return A2(
					rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetParent,
					_Utils_Tuple2(x_, y_),
					A2(rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$position, x_, y_));
			},
			A5(
				elm$json$Json$Decode$map4,
				F4(
					function (scrollLeft_, scrollTop_, offsetLeft_, offsetTop_) {
						return _Utils_Tuple2((x + offsetLeft_) - scrollLeft_, (y + offsetTop_) - scrollTop_);
					}),
				rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$scrollLeft,
				rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$scrollTop,
				rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetLeft,
				rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetTop));
	});
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$boundingArea = A4(
	elm$json$Json$Decode$map3,
	F3(
		function (_n0, width, height) {
			var x = _n0.a;
			var y = _n0.b;
			return {height: height, left: x, top: y, width: width};
		}),
	A2(rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$position, 0, 0),
	rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetWidth,
	rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$offsetHeight);
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$childNode = function (idx) {
	return elm$json$Json$Decode$at(
		_List_fromArray(
			[
				'childNodes',
				elm$core$String$fromInt(idx)
			]));
};
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$nextSibling = function (decoder) {
	return A2(elm$json$Json$Decode$field, 'nextSibling', decoder);
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$sizeDecoder = A3(
	elm$json$Json$Decode$map2,
	elm$core$Tuple$pair,
	A2(
		rundis$elm_bootstrap$Bootstrap$Dropdown$toggler,
		_List_fromArray(
			['target']),
		rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$boundingArea),
	A2(
		rundis$elm_bootstrap$Bootstrap$Dropdown$toggler,
		_List_fromArray(
			['target']),
		rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$nextSibling(
			A2(rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$childNode, 0, rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$boundingArea))));
var rundis$elm_bootstrap$Bootstrap$Dropdown$clickHandler = F2(
	function (toMsg, state) {
		var status = state.a.status;
		return A2(
			elm$json$Json$Decode$andThen,
			function (_n0) {
				var b = _n0.a;
				var m = _n0.b;
				return elm$json$Json$Decode$succeed(
					toMsg(
						rundis$elm_bootstrap$Bootstrap$Dropdown$State(
							{
								menuSize: m,
								status: rundis$elm_bootstrap$Bootstrap$Dropdown$nextStatus(status),
								toggleSize: b
							})));
			},
			rundis$elm_bootstrap$Bootstrap$Dropdown$sizeDecoder);
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$togglePrivate = F4(
	function (buttonOptions, children, toggleMsg, state) {
		return A2(
			elm$html$Html$button,
			_Utils_ap(
				rundis$elm_bootstrap$Bootstrap$Internal$Button$buttonAttributes(buttonOptions),
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('dropdown-toggle'),
						elm$html$Html$Attributes$type_('button'),
						A2(
						elm$html$Html$Events$on,
						'click',
						A2(rundis$elm_bootstrap$Bootstrap$Dropdown$clickHandler, toggleMsg, state))
					])),
			children);
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$toggle = F2(
	function (buttonOptions, children) {
		return rundis$elm_bootstrap$Bootstrap$Dropdown$DropdownToggle(
			A2(rundis$elm_bootstrap$Bootstrap$Dropdown$togglePrivate, buttonOptions, children));
	});
var author$project$Routes$Ls$buildActionDropdown = F3(
	function (model, actModel, entry) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Dropdown$dropdown,
			entry.dropdown,
			{
				items: _List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Ls$HistoryClicked(entry))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-history')
									]),
								_List_Nil),
								elm$html$Html$text(' History')
							])),
						rundis$elm_bootstrap$Bootstrap$Dropdown$divider,
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$anchorItem,
						_List_fromArray(
							[
								elm$html$Html$Attributes$href(
								'/get' + (author$project$Util$urlEncodePath(
									author$project$Util$joinPath(
										_List_fromArray(
											[
												actModel.self.path,
												author$project$Util$basename(entry.path)
											]))) + '?direct=yes')),
								elm$html$Html$Events$onClick(
								A2(author$project$Routes$Ls$ActionDropdownMsg, entry, rundis$elm_bootstrap$Bootstrap$Dropdown$initialState)),
								elm$html$Html$Attributes$disabled(
								!author$project$Routes$Ls$mayDownload(model))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-file-download')
									]),
								_List_Nil),
								elm$html$Html$text(' Download')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$anchorItem,
						_List_fromArray(
							[
								elm$html$Html$Attributes$href(
								'/get' + author$project$Util$urlEncodePath(
									author$project$Util$joinPath(
										_List_fromArray(
											[
												actModel.self.path,
												author$project$Util$basename(entry.path)
											])))),
								elm$html$Html$Events$onClick(
								A2(author$project$Routes$Ls$ActionDropdownMsg, entry, rundis$elm_bootstrap$Bootstrap$Dropdown$initialState)),
								elm$html$Html$Attributes$disabled(
								!author$project$Routes$Ls$mayDownload(model))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-eye')
									]),
								_List_Nil),
								elm$html$Html$text(' View')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$anchorItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Ls$ShareMsg(
									author$project$Modals$Share$show(
										_List_fromArray(
											[entry.path]))))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-share-alt')
									]),
								_List_Nil),
								elm$html$Html$text(' Share')
							])),
						rundis$elm_bootstrap$Bootstrap$Dropdown$divider,
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Ls$RemoveClicked(entry)),
								elm$html$Html$Attributes$disabled(
								!author$project$Routes$Ls$mayEdit(model))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-trash')
									]),
								_List_Nil),
								elm$html$Html$text(' Delete')
							])),
						rundis$elm_bootstrap$Bootstrap$Dropdown$divider,
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Ls$RenameMsg(
									author$project$Modals$Rename$show(entry.path))),
								elm$html$Html$Attributes$disabled(
								!author$project$Routes$Ls$mayEdit(model))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-file-signature')
									]),
								_List_Nil),
								elm$html$Html$text(' Rename')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Ls$MoveMsg(
									author$project$Modals$MoveCopy$show(entry.path))),
								elm$html$Html$Attributes$disabled(
								!author$project$Routes$Ls$mayEdit(model))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-arrow-right')
									]),
								_List_Nil),
								elm$html$Html$text(' Move')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Ls$CopyMsg(
									author$project$Modals$MoveCopy$show(entry.path))),
								elm$html$Html$Attributes$disabled(
								!author$project$Routes$Ls$mayEdit(model))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fa fa-md fa-copy')
									]),
								_List_Nil),
								elm$html$Html$text(' Copy')
							]))
					]),
				options: _List_Nil,
				toggleButton: A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$toggle,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Button$roleLink]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-ellipsis-h')
								]),
							_List_Nil)
						])),
				toggleMsg: author$project$Routes$Ls$ActionDropdownMsg(entry)
			});
	});
var author$project$Routes$Ls$formatPath = F2(
	function (model, entry) {
		var _n0 = model.isFiltered;
		if (_n0) {
			return A2(
				elm$core$String$join,
				'/',
				author$project$Util$splitPath(entry.path));
		} else {
			return author$project$Util$basename(entry.path);
		}
	});
var elm$html$Html$i = _VirtualDom_node('i');
var elm$html$Html$Attributes$checked = elm$html$Html$Attributes$boolProperty('checked');
var elm$html$Html$Events$targetChecked = A2(
	elm$json$Json$Decode$at,
	_List_fromArray(
		['target', 'checked']),
	elm$json$Json$Decode$bool);
var elm$html$Html$Events$onCheck = function (tagger) {
	return A2(
		elm$html$Html$Events$on,
		'change',
		A2(elm$json$Json$Decode$map, tagger, elm$html$Html$Events$targetChecked));
};
var author$project$Routes$Ls$makeCheckbox = F2(
	function (isChecked, msg) {
		return A2(
			elm$html$Html$div,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('checkbox')
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$label,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$input,
							_List_fromArray(
								[
									elm$html$Html$Attributes$type_('checkbox'),
									elm$html$Html$Events$onCheck(msg),
									elm$html$Html$Attributes$checked(isChecked)
								]),
							_List_Nil),
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('cr')
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$i,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('cr-icon fas fa-lg fa-check')
										]),
									_List_Nil)
								]))
						]))
				]));
	});
var author$project$Routes$Ls$readCheckedState = F2(
	function (model, path) {
		return A2(elm$core$Set$member, path, model.checked);
	});
var author$project$Routes$Ls$viewEntryIcon = function (entry) {
	var _n0 = entry.isDir;
	if (_n0) {
		return A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('fas fa-lg fa-folder text-xs-right file-list-icon')
				]),
			_List_Nil);
	} else {
		return A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('far fa-lg fa-file text-xs-right file-list-icon')
				]),
			_List_Nil);
	}
};
var author$project$Routes$Ls$PinClicked = F2(
	function (a, b) {
		return {$: 'PinClicked', a: a, b: b};
	});
var author$project$Routes$Ls$viewPinIcon = F2(
	function (isPinned, isExplicit) {
		var _n0 = _Utils_Tuple2(isPinned, isExplicit);
		if (_n0.a) {
			if (_n0.b) {
				return A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fa fa-map-marker'),
							elm$html$Html$Attributes$class('text-success')
						]),
					_List_Nil);
			} else {
				return A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fa fa-map-marker-alt'),
							elm$html$Html$Attributes$class('text-warning')
						]),
					_List_Nil);
			}
		} else {
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fa fa-times'),
						elm$html$Html$Attributes$class('text-danger')
					]),
				_List_Nil);
		}
	});
var author$project$Routes$Ls$viewPinButton = F2(
	function (model, entry) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Button$button,
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Button$roleLink,
					rundis$elm_bootstrap$Bootstrap$Button$attrs(
					_List_fromArray(
						[
							elm$html$Html$Attributes$disabled(
							!A2(elm$core$List$member, 'fs.edit', model.rights)),
							elm$html$Html$Events$onClick(
							A2(author$project$Routes$Ls$PinClicked, entry.path, !entry.isPinned))
						]))
				]),
			_List_fromArray(
				[
					A2(author$project$Routes$Ls$viewPinIcon, entry.isPinned, entry.isExplicit)
				]));
	});
var author$project$Util$formatLastModifiedOwner = F3(
	function (z, t, owner) {
		return A2(
			elm$html$Html$p,
			_List_Nil,
			_List_fromArray(
				[
					elm$html$Html$text(
					A2(author$project$Util$formatLastModified, z, t)),
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('text-muted')
						]),
					_List_fromArray(
						[
							elm$html$Html$text(' by ')
						])),
					elm$html$Html$text(owner)
				]));
	});
var author$project$Routes$Ls$entryToHtml = F4(
	function (model, actModel, zone, e) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Table$tr,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							author$project$Routes$Ls$makeCheckbox,
							A2(author$project$Routes$Ls$readCheckedState, actModel, e.path),
							author$project$Routes$Ls$CheckboxTick(e.path))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Attributes$class('icon-column')),
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Events$onClick(
								author$project$Routes$Ls$RowClicked(e)))
						]),
					_List_fromArray(
						[
							author$project$Routes$Ls$viewEntryIcon(e)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Events$onClick(
								author$project$Routes$Ls$RowClicked(e)))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$a,
							_List_fromArray(
								[
									elm$html$Html$Attributes$href('/view' + e.path)
								]),
							_List_fromArray(
								[
									elm$html$Html$text(
									A2(author$project$Routes$Ls$formatPath, actModel, e))
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Events$onClick(
								author$project$Routes$Ls$RowClicked(e)))
						]),
					_List_fromArray(
						[
							A3(author$project$Util$formatLastModifiedOwner, zone, e.lastModified, e.user)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Events$onClick(
								author$project$Routes$Ls$RowClicked(e)))
						]),
					_List_fromArray(
						[
							elm$html$Html$text(
							basti1302$elm_human_readable_filesize$Filesize$format(e.size))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(author$project$Routes$Ls$viewPinButton, model, e)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A3(author$project$Routes$Ls$buildActionDropdown, model, actModel, e)
						]))
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Table$simpleThead = function (cells) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Table$thead,
		_List_Nil,
		_List_fromArray(
			[
				A2(rundis$elm_bootstrap$Bootstrap$Table$tr, _List_Nil, cells)
			]));
};
var author$project$Routes$Ls$entriesToHtml = F3(
	function (model, zone, actModel) {
		return rundis$elm_bootstrap$Bootstrap$Table$table(
			{
				options: _List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Table$hover]),
				tbody: A2(
					rundis$elm_bootstrap$Bootstrap$Table$tbody,
					_List_Nil,
					A2(
						elm$core$List$map,
						A3(author$project$Routes$Ls$entryToHtml, model, actModel, zone),
						actModel.entries)),
				thead: rundis$elm_bootstrap$Bootstrap$Table$simpleThead(
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '5%'))
								]),
							_List_fromArray(
								[
									A2(
									author$project$Routes$Ls$makeCheckbox,
									A2(author$project$Routes$Ls$readCheckedState, actModel, ''),
									author$project$Routes$Ls$CheckboxTickAll)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '5%'))
								]),
							_List_fromArray(
								[
									elm$html$Html$text('')
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '37.5%'))
								]),
							_List_fromArray(
								[
									A3(author$project$Routes$Ls$buildSortControl, 'Name', actModel, author$project$Routes$Ls$Name)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '27.5%'))
								]),
							_List_fromArray(
								[
									A3(author$project$Routes$Ls$buildSortControl, 'Modified', actModel, author$project$Routes$Ls$ModTime)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '7.5%'))
								]),
							_List_fromArray(
								[
									A3(author$project$Routes$Ls$buildSortControl, 'Size', actModel, author$project$Routes$Ls$Size)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '10%'))
								]),
							_List_fromArray(
								[
									A3(author$project$Routes$Ls$buildSortControl, 'Pin', actModel, author$project$Routes$Ls$Pin)
								])),
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$th,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
									A2(elm$html$Html$Attributes$style, 'width', '5%'))
								]),
							_List_fromArray(
								[
									elm$html$Html$text('')
								]))
						]))
			});
	});
var rundis$elm_bootstrap$Bootstrap$Alert$dismissable = F2(
	function (dismissMsg, _n0) {
		var configRec = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Alert$Config(
			_Utils_update(
				configRec,
				{
					dismissable: elm$core$Maybe$Just(dismissMsg)
				}));
	});
var rundis$elm_bootstrap$Bootstrap$Alert$headingPrivate = F3(
	function (elemFn, attributes, children_) {
		return A2(
			elemFn,
			A2(
				elm$core$List$cons,
				elm$html$Html$Attributes$class('alert-header'),
				attributes),
			children_);
	});
var rundis$elm_bootstrap$Bootstrap$Alert$h4 = F2(
	function (attributes, children_) {
		return A3(rundis$elm_bootstrap$Bootstrap$Alert$headingPrivate, elm$html$Html$h4, attributes, children_);
	});
var author$project$Routes$Ls$showAlert = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Alert$view,
		model.alert,
		A2(
			rundis$elm_bootstrap$Bootstrap$Alert$children,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Alert$h4,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text('Oh, something went wrong! :(')
						])),
					elm$html$Html$text('The exact error was: ' + model.currError)
				]),
			rundis$elm_bootstrap$Bootstrap$Alert$danger(
				A2(rundis$elm_bootstrap$Bootstrap$Alert$dismissable, author$project$Routes$Ls$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$config))));
};
var author$project$Util$urlPrefixToString = function (url) {
	return function () {
		var _n0 = url.protocol;
		if (_n0.$ === 'Https') {
			return 'https://';
		} else {
			return 'http://';
		}
	}() + (url.host + (function () {
		var _n1 = url.port_;
		if (_n1.$ === 'Just') {
			var port_ = _n1.a;
			return ':' + elm$core$String$fromInt(port_);
		} else {
			return '';
		}
	}() + '/'));
};
var rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Outlined(rundis$elm_bootstrap$Bootstrap$Internal$Button$Primary));
var author$project$Routes$Ls$viewDownloadButton = F3(
	function (model, actModel, url) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Button$linkButton,
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
					rundis$elm_bootstrap$Bootstrap$Button$attrs(
					author$project$Routes$Ls$mayDownload(model) ? _List_fromArray(
						[
							elm$html$Html$Attributes$href(
							author$project$Util$urlPrefixToString(url) + ('get' + (author$project$Util$urlEncodePath(actModel.self.path) + '?direct=yes')))
						]) : _List_fromArray(
						[
							elm$html$Html$Attributes$class('text-muted'),
							A2(elm$html$Html$Attributes$style, 'opacity', '0.1')
						]))
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fas fa-download')
						]),
					_List_Nil),
					elm$html$Html$text(' Download')
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col4 = {$: 'Col4'};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs4 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col4);
var author$project$Routes$Ls$viewMetaRow = F2(
	function (key, value) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$xs4,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsLeft)
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted')
								]),
							_List_fromArray(
								[
									elm$html$Html$text(key)
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$xs8,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsRight)
						]),
					_List_fromArray(
						[value]))
				]));
	});
var author$project$Routes$Ls$viewViewButton = F3(
	function (model, actModel, url) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Button$linkButton,
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
					rundis$elm_bootstrap$Bootstrap$Button$attrs(
					author$project$Routes$Ls$mayDownload(model) ? _List_fromArray(
						[
							elm$html$Html$Attributes$href(
							author$project$Util$urlPrefixToString(url) + ('get' + author$project$Util$urlEncodePath(actModel.self.path)))
						]) : _List_fromArray(
						[
							elm$html$Html$Attributes$class('text-muted'),
							A2(elm$html$Html$Attributes$style, 'opacity', '0.1')
						]))
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fas fa-eye')
						]),
					_List_Nil),
					elm$html$Html$text(' View')
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$Roled = function (a) {
	return {$: 'Roled', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Internal$Role$Light = {$: 'Light'};
var rundis$elm_bootstrap$Bootstrap$ListGroup$light = rundis$elm_bootstrap$Bootstrap$Internal$ListGroup$Roled(rundis$elm_bootstrap$Bootstrap$Internal$Role$Light);
var author$project$Routes$Ls$viewSingleEntry = F3(
	function (model, actualModel, zone) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs2]),
					_List_Nil),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$xs8,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsCenter)
						]),
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$ListGroup$ul(
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$ListGroup$li,
									_List_Nil,
									_List_fromArray(
										[
											A2(
											author$project$Routes$Ls$viewMetaRow,
											'Path',
											elm$html$Html$text(actualModel.self.path))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$ListGroup$li,
									_List_Nil,
									_List_fromArray(
										[
											A2(
											author$project$Routes$Ls$viewMetaRow,
											'Size',
											elm$html$Html$text(
												basti1302$elm_human_readable_filesize$Filesize$format(actualModel.self.size)))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$ListGroup$li,
									_List_Nil,
									_List_fromArray(
										[
											A2(
											author$project$Routes$Ls$viewMetaRow,
											'Owner',
											elm$html$Html$text(actualModel.self.user))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$ListGroup$li,
									_List_Nil,
									_List_fromArray(
										[
											A2(
											author$project$Routes$Ls$viewMetaRow,
											'Last Modified',
											elm$html$Html$text(
												A2(author$project$Util$formatLastModified, zone, actualModel.self.lastModified)))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$ListGroup$li,
									_List_Nil,
									_List_fromArray(
										[
											A2(
											author$project$Routes$Ls$viewMetaRow,
											'Pinned',
											A2(author$project$Routes$Ls$viewPinButton, model, actualModel.self))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$ListGroup$li,
									_List_fromArray(
										[rundis$elm_bootstrap$Bootstrap$ListGroup$light]),
									_List_fromArray(
										[
											A3(author$project$Routes$Ls$viewDownloadButton, model, actualModel, model.url),
											elm$html$Html$text(' '),
											A3(author$project$Routes$Ls$viewViewButton, model, actualModel, model.url)
										]))
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs2]),
					_List_Nil)
				]));
	});
var elm$virtual_dom$VirtualDom$lazy3 = _VirtualDom_lazy3;
var elm$html$Html$Lazy$lazy3 = elm$virtual_dom$VirtualDom$lazy3;
var author$project$Routes$Ls$viewList = F2(
	function (model, zone) {
		var _n0 = model.state;
		switch (_n0.$) {
			case 'Failure':
				return A2(
					elm$html$Html$div,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text('Sorry, something did not work out as expected.')
						]));
			case 'Loading':
				return elm$html$Html$text('Loading...');
			default:
				var actualModel = _n0.a;
				var _n1 = actualModel.self.isDir;
				if (_n1) {
					return A2(
						elm$html$Html$div,
						_List_Nil,
						_List_fromArray(
							[
								author$project$Routes$Ls$showAlert(model),
								A4(elm$html$Html$Lazy$lazy3, author$project$Routes$Ls$entriesToHtml, model, zone, actualModel)
							]));
				} else {
					return A2(
						elm$html$Html$div,
						_List_Nil,
						_List_fromArray(
							[
								author$project$Routes$Ls$showAlert(model),
								A4(elm$html$Html$Lazy$lazy3, author$project$Routes$Ls$viewSingleEntry, model, actualModel, zone)
							]));
				}
		}
	});
var author$project$Routes$Ls$SearchInput = function (a) {
	return {$: 'SearchInput', a: a};
};
var author$project$Routes$Ls$viewSearchBox = function (model) {
	return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$view(
		A2(
			rundis$elm_bootstrap$Bootstrap$Form$InputGroup$attrs,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('stylish-input-group input-group')
				]),
			A2(
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$successors,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Form$InputGroup$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('input-group-addon')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$button,
								_List_Nil,
								_List_fromArray(
									[
										A2(
										elm$html$Html$span,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('fas fa-search fa-xs input-group-addon')
											]),
										_List_Nil)
									]))
							]))
					]),
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$config(
					rundis$elm_bootstrap$Bootstrap$Form$InputGroup$text(
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Search'),
								rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onInput(author$project$Routes$Ls$SearchInput),
										elm$html$Html$Attributes$value(
										author$project$Routes$Ls$searchQueryFromUrl(model.url))
									]))
							]))))));
};
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xl2 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XL, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col2);
var author$project$Routes$Ls$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Grid$row,
		_List_Nil,
		_List_fromArray(
			[
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$col,
				_List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg12]),
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$row,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
								_List_fromArray(
									[
										elm$html$Html$Attributes$id('main-header-row')
									]))
							]),
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl9]),
								_List_fromArray(
									[
										author$project$Routes$Ls$viewBreadcrumbs(model)
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl3]),
								_List_fromArray(
									[
										A2(elm$html$Html$Lazy$lazy, author$project$Routes$Ls$viewSearchBox, model)
									]))
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$row,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
								_List_fromArray(
									[
										elm$html$Html$Attributes$id('main-content-row')
									]))
							]),
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl10]),
								_List_fromArray(
									[
										A2(author$project$Routes$Ls$viewList, model, model.zone)
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$col,
								_List_fromArray(
									[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl2]),
								_List_fromArray(
									[
										A2(elm$html$Html$Lazy$lazy, author$project$Routes$Ls$viewActionList, model)
									]))
							]))
					]))
			]));
};
var author$project$Modals$RemoteAdd$ModalShow = {$: 'ModalShow'};
var author$project$Modals$RemoteAdd$show = author$project$Modals$RemoteAdd$ModalShow;
var author$project$Modals$RemoteFolders$ModalShow = function (a) {
	return {$: 'ModalShow', a: a};
};
var author$project$Modals$RemoteFolders$show = function (remote) {
	return author$project$Modals$RemoteFolders$ModalShow(remote);
};
var author$project$Routes$Remotes$AcceptPushToggled = F2(
	function (a, b) {
		return {$: 'AcceptPushToggled', a: a, b: b};
	});
var author$project$Util$viewToggleSwitch = F4(
	function (toMsg, message, isChecked, isDisabled) {
		return A2(
			elm$html$Html$span,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$label,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('toggle-switch'),
									elm$html$Html$Attributes$disabled(isDisabled),
									isDisabled ? elm$html$Html$Attributes$class('toggle-switch-disabled') : elm$html$Html$Attributes$class('')
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$input,
									_List_fromArray(
										[
											elm$html$Html$Attributes$type_('checkbox'),
											elm$html$Html$Events$onCheck(toMsg),
											elm$html$Html$Attributes$checked(isChecked),
											elm$html$Html$Attributes$disabled(isDisabled)
										]),
									_List_Nil),
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('toggle-slider toggle-round')
										]),
									_List_Nil)
								]))
						])),
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('text-muted')
						]),
					_List_fromArray(
						[
							elm$html$Html$text(' ' + message)
						]))
				]));
	});
var author$project$Routes$Remotes$viewAcceptPushToggle = F3(
	function (state, remote, isDisabled) {
		return A4(
			author$project$Util$viewToggleSwitch,
			author$project$Routes$Remotes$AcceptPushToggled(remote),
			'',
			state,
			isDisabled);
	});
var author$project$Modals$RemoteRemove$ModalShow = function (a) {
	return {$: 'ModalShow', a: a};
};
var author$project$Modals$RemoteRemove$show = function (name) {
	return author$project$Modals$RemoteRemove$ModalShow(name);
};
var author$project$Routes$Remotes$SyncClicked = function (a) {
	return {$: 'SyncClicked', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$AlignMenuRight = {$: 'AlignMenuRight'};
var rundis$elm_bootstrap$Bootstrap$Dropdown$alignMenuRight = rundis$elm_bootstrap$Bootstrap$Dropdown$AlignMenuRight;
var author$project$Routes$Remotes$viewActionDropdown = F2(
	function (model, remote) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Dropdown$dropdown,
			A2(
				elm$core$Maybe$withDefault,
				rundis$elm_bootstrap$Bootstrap$Dropdown$initialState,
				A2(elm$core$Dict$get, remote.name, model.actionDropdowns)),
			{
				items: _List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Remotes$SyncClicked(remote.name)),
								elm$html$Html$Attributes$disabled(
								(!remote.isAuthenticated) || (!A2(elm$core$List$member, 'fs.edit', model.rights)))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-sync-alt')
									]),
								_List_Nil),
								elm$html$Html$text(' Sync')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$anchorItem,
						_List_fromArray(
							[
								elm$html$Html$Attributes$disabled(
								(!remote.isAuthenticated) || (!A2(elm$core$List$member, 'fs.view', model.rights))),
								remote.isAuthenticated ? elm$html$Html$Attributes$href(
								'/diff/' + elm$url$Url$percentEncode(remote.name)) : elm$html$Html$Attributes$class('text-muted')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-search-minus')
									]),
								_List_Nil),
								elm$html$Html$text(' Diff')
							])),
						rundis$elm_bootstrap$Bootstrap$Dropdown$divider,
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								author$project$Routes$Remotes$RemoteRemoveMsg(
									author$project$Modals$RemoteRemove$show(remote.name))),
								elm$html$Html$Attributes$disabled(
								!A2(elm$core$List$member, 'remotes.edit', model.rights))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('text-danger')
									]),
								_List_fromArray(
									[
										A2(
										elm$html$Html$span,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('fas fa-md fa-times')
											]),
										_List_Nil),
										elm$html$Html$text(' Remove')
									]))
							]))
					]),
				options: _List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Dropdown$alignMenuRight]),
				toggleButton: A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$toggle,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Button$roleLink]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-ellipsis-h')
								]),
							_List_Nil)
						])),
				toggleMsg: author$project$Routes$Remotes$ActionDropdownMsg(remote.name)
			});
	});
var author$project$Routes$Remotes$AutoUpdateToggled = F2(
	function (a, b) {
		return {$: 'AutoUpdateToggled', a: a, b: b};
	});
var author$project$Routes$Remotes$viewAutoUpdatesIcon = F3(
	function (state, remote, isDisabled) {
		return A4(
			author$project$Util$viewToggleSwitch,
			author$project$Routes$Remotes$AutoUpdateToggled(remote),
			'',
			state,
			isDisabled);
	});
var author$project$Routes$Remotes$ConflictStrategyToggled = F2(
	function (a, b) {
		return {$: 'ConflictStrategyToggled', a: a, b: b};
	});
var author$project$Routes$Remotes$conflictStrategyToIconName = F2(
	function (model, strategy) {
		conflictStrategyToIconName:
		while (true) {
			switch (strategy) {
				case '':
					if (model.self.defaultConflictStrategy === '') {
						return 'fa-question text-muted';
					} else {
						var $temp$model = model,
							$temp$strategy = model.self.defaultConflictStrategy;
						model = $temp$model;
						strategy = $temp$strategy;
						continue conflictStrategyToIconName;
					}
				case 'ignore':
					return 'fa-eject';
				case 'marker':
					return 'fa-marker';
				case 'embrace':
					return 'fa-handshake';
				default:
					return 'fa-question';
			}
		}
	});
var rundis$elm_bootstrap$Bootstrap$Dropdown$Attrs = function (a) {
	return {$: 'Attrs', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Dropdown$attrs = function (attrs_) {
	return rundis$elm_bootstrap$Bootstrap$Dropdown$Attrs(attrs_);
};
var author$project$Routes$Remotes$viewConflictDropdown = F3(
	function (model, remote, isDisabled) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Dropdown$dropdown,
			A2(
				elm$core$Maybe$withDefault,
				rundis$elm_bootstrap$Bootstrap$Dropdown$initialState,
				A2(elm$core$Dict$get, remote.name, model.conflictDropdowns)),
			{
				items: _List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Routes$Remotes$ConflictStrategyToggled, remote, 'ignore')),
								elm$html$Html$Attributes$disabled(isDisabled)
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-eject')
									]),
								_List_Nil),
								elm$html$Html$text(' Ignore')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Routes$Remotes$ConflictStrategyToggled, remote, 'marker')),
								elm$html$Html$Attributes$disabled(isDisabled)
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-marker')
									]),
								_List_Nil),
								elm$html$Html$text(' Marker')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Routes$Remotes$ConflictStrategyToggled, remote, 'embrace')),
								elm$html$Html$Attributes$disabled(isDisabled)
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-handshake')
									]),
								_List_Nil),
								elm$html$Html$text(' Embrace')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Routes$Remotes$ConflictStrategyToggled, remote, '')),
								elm$html$Html$Attributes$disabled(isDisabled)
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-eraser')
									]),
								_List_Nil),
								elm$html$Html$text(' Default')
							]))
					]),
				options: _List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Dropdown$alignMenuRight,
						rundis$elm_bootstrap$Bootstrap$Dropdown$attrs(
						_List_fromArray(
							[
								elm$html$Html$Attributes$disabled(isDisabled)
							]))
					]),
				toggleButton: A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$toggle,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$roleLink,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$disabled(isDisabled)
								]))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas'),
									elm$html$Html$Attributes$class(
									A2(author$project$Routes$Remotes$conflictStrategyToIconName, model, remote.conflictStrategy))
								]),
							_List_Nil)
						])),
				toggleMsg: author$project$Routes$Remotes$ConflictDropdownMsg(remote.name)
			});
	});
var elm$core$List$intersperse = F2(
	function (sep, xs) {
		if (!xs.b) {
			return _List_Nil;
		} else {
			var hd = xs.a;
			var tl = xs.b;
			var step = F2(
				function (x, rest) {
					return A2(
						elm$core$List$cons,
						sep,
						A2(elm$core$List$cons, x, rest));
				});
			var spersed = A3(elm$core$List$foldr, step, _List_Nil, tl);
			return A2(elm$core$List$cons, hd, spersed);
		}
	});
var author$project$Routes$Remotes$viewFullFingerprint = function (fingerprint) {
	return A2(
		elm$html$Html$span,
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('fingerprint')
			]),
		A2(
			elm$core$List$intersperse,
			A2(
				elm$html$Html$span,
				_List_Nil,
				_List_fromArray(
					[
						elm$html$Html$text(':'),
						A2(elm$html$Html$br, _List_Nil, _List_Nil)
					])),
			A2(
				elm$core$List$map,
				function (t) {
					return A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('text-muted')
							]),
						_List_fromArray(
							[
								elm$html$Html$text(t)
							]));
				},
				A2(elm$core$String$split, ':', fingerprint))));
};
var author$project$Routes$Remotes$viewRemoteState = F2(
	function (model, remote) {
		return remote.isAuthenticated ? (remote.isOnline ? A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('fas fa-md fa-circle text-success')
				]),
			_List_Nil) : A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('text-warning')
				]),
			_List_fromArray(
				[
					elm$html$Html$text(
					A2(author$project$Util$formatLastModified, model.zone, remote.lastSeen))
				]))) : A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('text-danger')
				]),
			_List_fromArray(
				[
					elm$html$Html$text('not authenticated')
				]));
	});
var author$project$Routes$Remotes$viewRemote = F2(
	function (model, remote) {
		var isDisabled = !A2(elm$core$List$member, 'remotes.edit', model.rights);
		return A2(
			rundis$elm_bootstrap$Bootstrap$Table$tr,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-lg fa-user-circle text-xs-right')
								]),
							_List_Nil)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text(' ' + remote.name)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(author$project$Routes$Remotes$viewRemoteState, model, remote)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted')
								]),
							_List_fromArray(
								[
									author$project$Routes$Remotes$viewFullFingerprint(remote.fingerprint)
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A3(author$project$Routes$Remotes$viewAutoUpdatesIcon, remote.acceptAutoUpdates, remote, isDisabled)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A3(author$project$Routes$Remotes$viewAcceptPushToggle, remote.acceptPush, remote, isDisabled)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A3(author$project$Routes$Remotes$viewConflictDropdown, model, remote, isDisabled)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Button$button,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Button$roleLink,
									rundis$elm_bootstrap$Bootstrap$Button$attrs(
									_List_fromArray(
										[
											elm$html$Html$Events$onClick(
											author$project$Routes$Remotes$RemoteFolderMsg(
												author$project$Modals$RemoteFolders$show(remote))),
											elm$html$Html$Attributes$disabled(isDisabled)
										]))
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_Nil,
									_List_fromArray(
										[
											function () {
											var _n0 = elm$core$List$length(remote.folders);
											if (!_n0) {
												return A2(
													elm$html$Html$span,
													_List_fromArray(
														[
															elm$html$Html$Attributes$class('fas fa-xs fa-asterisk')
														]),
													_List_Nil);
											} else {
												var n = _n0;
												return elm$html$Html$text(
													elm$core$String$fromInt(n));
											}
										}()
										]))
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Attributes$class('text-right'))
						]),
					_List_fromArray(
						[
							A2(author$project$Routes$Remotes$viewActionDropdown, model, remote)
						]))
				]));
	});
var author$project$Routes$Remotes$viewRemoteList = F2(
	function (model, remotes) {
		return rundis$elm_bootstrap$Bootstrap$Table$table(
			{
				options: _List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Table$hover,
						rundis$elm_bootstrap$Bootstrap$Table$attr(
						elm$html$Html$Attributes$class('borderless-table'))
					]),
				tbody: A2(
					rundis$elm_bootstrap$Bootstrap$Table$tbody,
					_List_Nil,
					A2(
						elm$core$List$map,
						author$project$Routes$Remotes$viewRemote(model),
						remotes)),
				thead: A2(
					rundis$elm_bootstrap$Bootstrap$Table$thead,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$tr,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '5%'))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '20%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Name')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '20%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Online')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '30%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Fingerprint')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '10%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Auto Update')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '10%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('May Push')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '10%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Conflicts')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '10%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Folders')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '5%'))
										]),
									_List_Nil)
								]))
						]))
			});
	});
var author$project$Routes$Remotes$viewRemoteListContainer = F2(
	function (model, remotes) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg1,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12, rundis$elm_bootstrap$Bootstrap$Grid$Col$lg10]),
					_List_fromArray(
						[
							A2(author$project$Util$viewAlert, author$project$Routes$Remotes$AlertMsg, model.alert),
							A2(author$project$Routes$Remotes$viewRemoteList, model, remotes),
							A2(
							elm$html$Html$div,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-left')
								]),
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Button$button,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Button$roleLink,
											rundis$elm_bootstrap$Bootstrap$Button$attrs(
											_List_fromArray(
												[
													elm$html$Html$Events$onClick(
													author$project$Routes$Remotes$RemoteAddMsg(author$project$Modals$RemoteAdd$show)),
													elm$html$Html$Attributes$disabled(
													!A2(elm$core$List$member, 'remotes.edit', model.rights))
												]))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('fas fa-lg fa-plus')
												]),
											_List_Nil),
											elm$html$Html$text(' Add new')
										]))
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$lg1,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('d-none d-lg-block')
								]))
						]),
					_List_Nil)
				]));
	});
var author$project$Routes$Remotes$viewMetaRow = F2(
	function (key, value) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$xs4,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsLeft)
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted')
								]),
							_List_fromArray(
								[
									elm$html$Html$text(key)
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Grid$Col$xs8,
							rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsRight)
						]),
					_List_fromArray(
						[value]))
				]));
	});
var author$project$Routes$Remotes$viewSelf = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Grid$row,
		_List_Nil,
		_List_fromArray(
			[
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$col,
				_List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2,
						rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('d-none d-lg-block')
							]))
					]),
				_List_Nil),
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$col,
				_List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12,
						rundis$elm_bootstrap$Bootstrap$Grid$Col$lg8,
						rundis$elm_bootstrap$Bootstrap$Grid$Col$textAlign(rundis$elm_bootstrap$Bootstrap$Text$alignXsCenter)
					]),
				_List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$ListGroup$ul(
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$ListGroup$li,
								_List_Nil,
								_List_fromArray(
									[
										A2(
										author$project$Routes$Remotes$viewMetaRow,
										'Name',
										elm$html$Html$text(model.self.self.name))
									])),
								A2(
								rundis$elm_bootstrap$Bootstrap$ListGroup$li,
								_List_Nil,
								_List_fromArray(
									[
										A2(
										author$project$Routes$Remotes$viewMetaRow,
										'Fingerprint',
										author$project$Routes$Remotes$viewFullFingerprint(model.self.self.fingerprint))
									]))
							]))
					])),
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$col,
				_List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Grid$Col$lg2,
						rundis$elm_bootstrap$Bootstrap$Grid$Col$attrs(
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('d-none d-lg-block')
							]))
					]),
				_List_Nil)
			]));
};
var author$project$Routes$Remotes$view = function (model) {
	var _n0 = model.state;
	switch (_n0.$) {
		case 'Loading':
			return elm$html$Html$text('Still loading');
		case 'Failure':
			var err = _n0.a;
			return elm$html$Html$text('Failed to load remote list: ' + err);
		default:
			var remotes = _n0.a;
			return A2(
				rundis$elm_bootstrap$Bootstrap$Grid$row,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$col,
						_List_fromArray(
							[rundis$elm_bootstrap$Bootstrap$Grid$Col$lg12]),
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$id('main-header-row')
											]))
									]),
								_List_Nil),
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$id('main-content-row')
											]))
									]),
								_List_fromArray(
									[
										A2(
										rundis$elm_bootstrap$Bootstrap$Grid$col,
										_List_fromArray(
											[rundis$elm_bootstrap$Bootstrap$Grid$Col$xl10]),
										_List_fromArray(
											[
												A2(
												elm$html$Html$h4,
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('text-center text-muted')
													]),
												_List_fromArray(
													[
														elm$html$Html$text('Own data')
													])),
												A2(elm$html$Html$br, _List_Nil, _List_Nil),
												author$project$Routes$Remotes$viewSelf(model),
												A2(elm$html$Html$br, _List_Nil, _List_Nil),
												A2(elm$html$Html$br, _List_Nil, _List_Nil),
												A2(elm$html$Html$br, _List_Nil, _List_Nil),
												A2(elm$html$Html$br, _List_Nil, _List_Nil),
												A2(
												elm$html$Html$h4,
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('text-center text-muted')
													]),
												_List_fromArray(
													[
														elm$html$Html$text('Other remotes')
													])),
												A2(elm$html$Html$br, _List_Nil, _List_Nil),
												A2(author$project$Routes$Remotes$viewRemoteListContainer, model, remotes)
											]))
									]))
							]))
					]));
	}
};
var author$project$Main$viewCurrentRoute = F2(
	function (model, viewState) {
		var _n0 = viewState.currentView;
		switch (_n0.$) {
			case 'ViewList':
				return A2(
					elm$html$Html$map,
					author$project$Main$ListMsg,
					author$project$Routes$Ls$view(viewState.listState));
			case 'ViewCommits':
				return A2(
					elm$html$Html$map,
					author$project$Main$CommitsMsg,
					author$project$Routes$Commits$view(viewState.commitsState));
			case 'ViewDeletedFiles':
				return A2(
					elm$html$Html$map,
					author$project$Main$DeletedFilesMsg,
					author$project$Routes$DeletedFiles$view(viewState.deletedFilesState));
			case 'ViewRemotes':
				return A2(
					elm$html$Html$map,
					author$project$Main$RemotesMsg,
					author$project$Routes$Remotes$view(viewState.remoteState));
			case 'ViewDiff':
				return A2(
					elm$html$Html$map,
					author$project$Main$DiffMsg,
					author$project$Routes$Diff$view(viewState.diffState));
			default:
				return elm$html$Html$text('You seem to have hit a route that does not exist...');
		}
	});
var author$project$Main$viewOfflineMarker = A2(
	elm$html$Html$div,
	_List_fromArray(
		[
			elm$html$Html$Attributes$class('row h-100')
		]),
	_List_fromArray(
		[
			A2(
			elm$html$Html$div,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('col-12 my-auto text-center w-100 text-muted')
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fas fa-4x fa-fw logo-failure fa-plug')
						]),
					_List_Nil),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					elm$html$Html$text('It seems that we have lost connection to the server.'),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					elm$html$Html$text('This application will go into a working state again when we have a connection again.')
				]))
		]));
var elm$html$Html$hr = _VirtualDom_node('hr');
var author$project$Main$viewSidebarBottom = function (model) {
	return A2(
		elm$html$Html$div,
		_List_fromArray(
			[
				elm$html$Html$Attributes$id('sidebar-bottom'),
				elm$html$Html$Attributes$class('d-none d-lg-block')
			]),
		_List_fromArray(
			[
				A2(elm$html$Html$hr, _List_Nil, _List_Nil),
				A2(
				elm$html$Html$p,
				_List_fromArray(
					[
						elm$html$Html$Attributes$id('sidebar-bottom-text'),
						elm$html$Html$Attributes$class('text-muted')
					]),
				_List_fromArray(
					[
						A2(
						elm$html$Html$span,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text('Powered by '),
								A2(
								elm$html$Html$a,
								_List_fromArray(
									[
										elm$html$Html$Attributes$href('https://github.com/sahib/brig')
									]),
								_List_fromArray(
									[
										elm$html$Html$text('brig')
									]))
							]))
					]))
			]));
};
var author$project$Main$LogoutSubmit = function (a) {
	return {$: 'LogoutSubmit', a: a};
};
var author$project$Main$hasRight = F3(
	function (viewState, right, elements) {
		return A2(elm$core$List$member, right, viewState.rights) ? elements : _List_Nil;
	});
var author$project$Main$viewToString = function (v) {
	switch (v.$) {
		case 'ViewList':
			return '/view';
		case 'ViewCommits':
			return '/log';
		case 'ViewRemotes':
			return '/remotes';
		case 'ViewDeletedFiles':
			return '/deleted';
		case 'ViewDiff':
			return '/Diff';
		default:
			return '/nothing';
	}
};
var author$project$Main$viewSidebarItems = F2(
	function (model, viewState) {
		var isActiveClass = function (v) {
			return _Utils_eq(v, viewState.currentView) ? elm$html$Html$Attributes$class('nav-link active') : elm$html$Html$Attributes$class('nav-link');
		};
		return A2(
			elm$html$Html$ul,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('flex-column navbar-nav w-100 text-left')
				]),
			_Utils_ap(
				A3(
					author$project$Main$hasRight,
					viewState,
					'fs.view',
					_List_fromArray(
						[
							A2(
							elm$html$Html$li,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('nav-item')
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$a,
									_List_fromArray(
										[
											isActiveClass(author$project$Main$ViewList),
											elm$html$Html$Attributes$href(
											author$project$Main$viewToString(author$project$Main$ViewList))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_Nil,
											_List_fromArray(
												[
													elm$html$Html$text('Files')
												]))
										]))
								]))
						])),
				_Utils_ap(
					A3(
						author$project$Main$hasRight,
						viewState,
						'fs.view',
						_List_fromArray(
							[
								A2(
								elm$html$Html$li,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('nav-item')
									]),
								_List_fromArray(
									[
										A2(
										elm$html$Html$a,
										_List_fromArray(
											[
												isActiveClass(author$project$Main$ViewCommits),
												elm$html$Html$Attributes$href(
												author$project$Main$viewToString(author$project$Main$ViewCommits))
											]),
										_List_fromArray(
											[
												A2(
												elm$html$Html$span,
												_List_Nil,
												_List_fromArray(
													[
														elm$html$Html$text('Changelog')
													]))
											]))
									]))
							])),
					_Utils_ap(
						A3(
							author$project$Main$hasRight,
							viewState,
							'fs.view',
							_List_fromArray(
								[
									A2(
									elm$html$Html$li,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('nav-item')
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$a,
											_List_fromArray(
												[
													isActiveClass(author$project$Main$ViewDeletedFiles),
													elm$html$Html$Attributes$href(
													author$project$Main$viewToString(author$project$Main$ViewDeletedFiles))
												]),
											_List_fromArray(
												[
													A2(
													elm$html$Html$span,
													_List_Nil,
													_List_fromArray(
														[
															elm$html$Html$text('Trashbin')
														]))
												]))
										]))
								])),
						_Utils_ap(
							A3(
								author$project$Main$hasRight,
								viewState,
								'remotes.view',
								_List_fromArray(
									[
										A2(
										elm$html$Html$li,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('nav-item')
											]),
										_List_fromArray(
											[
												A2(
												elm$html$Html$a,
												_List_fromArray(
													[
														isActiveClass(author$project$Main$ViewRemotes),
														elm$html$Html$Attributes$href(
														author$project$Main$viewToString(author$project$Main$ViewRemotes))
													]),
												_List_fromArray(
													[
														A2(
														elm$html$Html$span,
														_List_Nil,
														_List_fromArray(
															[
																elm$html$Html$text('Remotes')
															]))
													]))
											]))
									])),
							viewState.isAnon ? _List_fromArray(
								[
									A2(
									elm$html$Html$li,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('nav-item')
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$a,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('nav-link pl-0'),
													elm$html$Html$Attributes$href('#'),
													elm$html$Html$Events$onClick(
													author$project$Main$LogoutSubmit(false))
												]),
											_List_fromArray(
												[
													A2(
													elm$html$Html$span,
													_List_Nil,
													_List_fromArray(
														[
															elm$html$Html$text('Login page')
														]))
												]))
										]))
								]) : _List_fromArray(
								[
									A2(
									elm$html$Html$li,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('nav-item')
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$a,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('nav-link pl-0'),
													elm$html$Html$Attributes$href('#'),
													elm$html$Html$Events$onClick(
													author$project$Main$LogoutSubmit(true))
												]),
											_List_fromArray(
												[
													A2(
													elm$html$Html$span,
													_List_Nil,
													_List_fromArray(
														[
															elm$html$Html$text('Logout »' + (viewState.loginName + '«'))
														]))
												]))
										]))
								]))))));
	});
var author$project$Modals$History$ModalClose = {$: 'ModalClose'};
var author$project$Modals$History$PinClicked = F3(
	function (a, b, c) {
		return {$: 'PinClicked', a: a, b: b, c: c};
	});
var author$project$Modals$History$ResetClicked = F2(
	function (a, b) {
		return {$: 'ResetClicked', a: a, b: b};
	});
var author$project$Modals$History$joinChanges = function (changes) {
	return A2(
		elm$core$List$intersperse,
		elm$html$Html$text(', '),
		changes);
};
var author$project$Modals$History$viewChangeColor = function (change) {
	switch (change) {
		case 'added':
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('text-success')
					]),
				_List_fromArray(
					[
						elm$html$Html$text(change)
					]));
		case 'modified':
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('text-warning')
					]),
				_List_fromArray(
					[
						elm$html$Html$text(change)
					]));
		case 'removed':
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('text-danger')
					]),
				_List_fromArray(
					[
						elm$html$Html$text(change)
					]));
		case 'moved':
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('text-info')
					]),
				_List_fromArray(
					[
						elm$html$Html$text(change)
					]));
		default:
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('text-muted')
					]),
				_List_fromArray(
					[
						elm$html$Html$text(change)
					]));
	}
};
var author$project$Modals$History$viewChangeSet = function (change) {
	var changes = A2(
		elm$core$List$map,
		author$project$Modals$History$viewChangeColor,
		A2(elm$core$String$split, '|', change));
	return A2(
		elm$html$Html$span,
		_List_Nil,
		author$project$Modals$History$joinChanges(changes));
};
var author$project$Modals$History$viewPinIcon = F2(
	function (isPinned, isExplicit) {
		var _n0 = _Utils_Tuple2(isPinned, isExplicit);
		if (_n0.a) {
			if (_n0.b) {
				return A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fa fa-map-marker'),
							elm$html$Html$Attributes$class('text-success')
						]),
					_List_Nil);
			} else {
				return A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fa fa-map-marker-alt'),
							elm$html$Html$Attributes$class('text-warning')
						]),
					_List_Nil);
			}
		} else {
			return A2(
				elm$html$Html$span,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('fa fa-times'),
						elm$html$Html$Attributes$class('text-danger')
					]),
				_List_Nil);
		}
	});
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$ButtonItem = function (a) {
	return {$: 'ButtonItem', a: a};
};
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$button = F2(
	function (options, children) {
		return rundis$elm_bootstrap$Bootstrap$ButtonGroup$ButtonItem(
			A2(rundis$elm_bootstrap$Bootstrap$Button$button, options, children));
	});
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$GroupItem = function (a) {
	return {$: 'GroupItem', a: a};
};
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$applyModifier = F2(
	function (modifier, options) {
		switch (modifier.$) {
			case 'Size':
				var size = modifier.a;
				return _Utils_update(
					options,
					{
						size: elm$core$Maybe$Just(size)
					});
			case 'Vertical':
				return _Utils_update(
					options,
					{vertical: true});
			default:
				var attrs_ = modifier.a;
				return _Utils_update(
					options,
					{
						attributes: _Utils_ap(options.attributes, attrs_)
					});
		}
	});
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$defaultOptions = {attributes: _List_Nil, size: elm$core$Maybe$Nothing, vertical: false};
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$groupAttributes = F2(
	function (toggle, modifiers) {
		var options = A3(elm$core$List$foldl, rundis$elm_bootstrap$Bootstrap$ButtonGroup$applyModifier, rundis$elm_bootstrap$Bootstrap$ButtonGroup$defaultOptions, modifiers);
		return _Utils_ap(
			_List_fromArray(
				[
					A2(elm$html$Html$Attributes$attribute, 'role', 'group'),
					elm$html$Html$Attributes$classList(
					_List_fromArray(
						[
							_Utils_Tuple2('btn-group', true),
							_Utils_Tuple2('btn-group-toggle', toggle),
							_Utils_Tuple2('btn-group-vertical', options.vertical)
						])),
					A2(elm$html$Html$Attributes$attribute, 'data-toggle', 'buttons')
				]),
			_Utils_ap(
				function () {
					var _n0 = A2(elm$core$Maybe$andThen, rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption, options.size);
					if (_n0.$ === 'Just') {
						var s = _n0.a;
						return _List_fromArray(
							[
								elm$html$Html$Attributes$class('btn-group-' + s)
							]);
					} else {
						return _List_Nil;
					}
				}(),
				options.attributes));
	});
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$buttonGroupItem = F2(
	function (options, items) {
		return rundis$elm_bootstrap$Bootstrap$ButtonGroup$GroupItem(
			A2(
				elm$html$Html$div,
				A2(rundis$elm_bootstrap$Bootstrap$ButtonGroup$groupAttributes, false, options),
				A2(
					elm$core$List$map,
					function (_n0) {
						var elem = _n0.a;
						return elem;
					},
					items)));
	});
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$renderGroup = function (_n0) {
	var elem = _n0.a;
	return elem;
};
var rundis$elm_bootstrap$Bootstrap$ButtonGroup$buttonGroup = F2(
	function (options, items) {
		return rundis$elm_bootstrap$Bootstrap$ButtonGroup$renderGroup(
			A2(rundis$elm_bootstrap$Bootstrap$ButtonGroup$buttonGroupItem, options, items));
	});
var rundis$elm_bootstrap$Bootstrap$Grid$Col$xs9 = A2(rundis$elm_bootstrap$Bootstrap$Grid$Internal$width, rundis$elm_bootstrap$Bootstrap$General$Internal$XS, rundis$elm_bootstrap$Bootstrap$Grid$Internal$Col9);
var author$project$Modals$History$viewHistoryEntry = F3(
	function (model, isFirst, entry) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs9]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$p,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text(entry.path),
									A2(elm$html$Html$br, _List_Nil, _List_Nil),
									author$project$Modals$History$viewChangeSet(entry.change),
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('text-muted')
										]),
									_List_fromArray(
										[
											elm$html$Html$text(' at ')
										])),
									elm$html$Html$text(
									A2(author$project$Util$formatLastModified, elm$time$Time$utc, entry.head.date)),
									elm$html$Html$text(': '),
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('text-muted')
										]),
									_List_fromArray(
										[
											elm$html$Html$text(entry.head.msg)
										]))
								]))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs3]),
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$ButtonGroup$buttonGroup,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$ButtonGroup$button,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
											rundis$elm_bootstrap$Bootstrap$Button$attrs(
											_List_fromArray(
												[
													elm$html$Html$Events$onClick(
													A2(author$project$Modals$History$ResetClicked, entry.path, entry.head.hash)),
													elm$html$Html$Attributes$disabled(isFirst)
												]))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('Revert')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$ButtonGroup$button,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
											rundis$elm_bootstrap$Bootstrap$Button$attrs(
											_List_fromArray(
												[
													elm$html$Html$Attributes$disabled(
													!A2(elm$core$List$member, 'fs.edit', model.rights)),
													elm$html$Html$Events$onClick(
													A3(author$project$Modals$History$PinClicked, entry.path, entry.head.hash, !entry.isPinned))
												]))
										]),
									_List_fromArray(
										[
											A2(author$project$Modals$History$viewPinIcon, entry.isPinned, entry.isExplicit)
										]))
								]))
						]))
				]));
	});
var author$project$Modals$History$viewHistoryEntries = F2(
	function (model, entries) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Grid$row,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Grid$col,
					_List_Nil,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$ListGroup$ul(
							A2(
								elm$core$List$indexedMap,
								F2(
									function (idx, e) {
										return A2(
											rundis$elm_bootstrap$Bootstrap$ListGroup$li,
											_List_Nil,
											_List_fromArray(
												[
													A3(author$project$Modals$History$viewHistoryEntry, model, !idx, e)
												]));
									}),
								entries))
						]))
				]));
	});
var author$project$Util$buildAlert = F5(
	function (visibility, msg, severity, title, message) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Alert$view,
			visibility,
			A2(
				rundis$elm_bootstrap$Bootstrap$Alert$children,
				_List_fromArray(
					[
						(elm$core$String$length(title) > 0) ? A2(
						rundis$elm_bootstrap$Bootstrap$Alert$h4,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text(title)
							])) : elm$html$Html$text(''),
						elm$html$Html$text(message)
					]),
				severity(
					A2(rundis$elm_bootstrap$Bootstrap$Alert$dismissableWithAnimation, msg, rundis$elm_bootstrap$Bootstrap$Alert$config))));
	});
var author$project$Modals$History$viewHistory = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					function () {
					var _n0 = model.history;
					if (_n0.$ === 'Nothing') {
						return elm$html$Html$text('');
					} else {
						var result = _n0.a;
						if (result.$ === 'Ok') {
							var entries = result.a;
							return A2(author$project$Modals$History$viewHistoryEntries, model, entries);
						} else {
							var err = result.a;
							return A5(
								author$project$Util$buildAlert,
								model.alert,
								author$project$Modals$History$AlertMsg,
								rundis$elm_bootstrap$Bootstrap$Alert$danger,
								'Oh no!',
								'Could not read history: ' + author$project$Util$httpErrorToString(err));
						}
					}
				}()
				]))
		]);
};
var rundis$elm_bootstrap$Bootstrap$Modal$Body = function (a) {
	return {$: 'Body', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Modal$Config = function (a) {
	return {$: 'Config', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Modal$body = F3(
	function (attributes, children, _n0) {
		var conf = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Modal$Config(
			_Utils_update(
				conf,
				{
					body: elm$core$Maybe$Just(
						rundis$elm_bootstrap$Bootstrap$Modal$Body(
							{attributes: attributes, children: children}))
				}));
	});
var rundis$elm_bootstrap$Bootstrap$Modal$config = function (closeMsg) {
	return rundis$elm_bootstrap$Bootstrap$Modal$Config(
		{
			body: elm$core$Maybe$Nothing,
			closeMsg: closeMsg,
			footer: elm$core$Maybe$Nothing,
			header: elm$core$Maybe$Nothing,
			options: {centered: true, hideOnBackdropClick: true, modalSize: elm$core$Maybe$Nothing},
			withAnimation: elm$core$Maybe$Nothing
		});
};
var rundis$elm_bootstrap$Bootstrap$Modal$Footer = function (a) {
	return {$: 'Footer', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Modal$footer = F3(
	function (attributes, children, _n0) {
		var conf = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Modal$Config(
			_Utils_update(
				conf,
				{
					footer: elm$core$Maybe$Just(
						rundis$elm_bootstrap$Bootstrap$Modal$Footer(
							{attributes: attributes, children: children}))
				}));
	});
var rundis$elm_bootstrap$Bootstrap$Modal$Header = function (a) {
	return {$: 'Header', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Modal$header = F3(
	function (attributes, children, _n0) {
		var conf = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Modal$Config(
			_Utils_update(
				conf,
				{
					header: elm$core$Maybe$Just(
						rundis$elm_bootstrap$Bootstrap$Modal$Header(
							{attributes: attributes, children: children}))
				}));
	});
var rundis$elm_bootstrap$Bootstrap$Modal$StartClose = {$: 'StartClose'};
var rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated = rundis$elm_bootstrap$Bootstrap$Modal$StartClose;
var rundis$elm_bootstrap$Bootstrap$Modal$large = function (_n0) {
	var conf = _n0.a;
	var options = conf.options;
	return rundis$elm_bootstrap$Bootstrap$Modal$Config(
		_Utils_update(
			conf,
			{
				options: _Utils_update(
					options,
					{
						modalSize: elm$core$Maybe$Just(rundis$elm_bootstrap$Bootstrap$General$Internal$LG)
					})
			}));
};
var elm$html$Html$Attributes$tabindex = function (n) {
	return A2(
		_VirtualDom_attribute,
		'tabIndex',
		elm$core$String$fromInt(n));
};
var rundis$elm_bootstrap$Bootstrap$Modal$getCloseMsg = function (config_) {
	var _n0 = config_.withAnimation;
	if (_n0.$ === 'Just') {
		var animationMsg = _n0.a;
		return animationMsg(rundis$elm_bootstrap$Bootstrap$Modal$StartClose);
	} else {
		return config_.closeMsg;
	}
};
var rundis$elm_bootstrap$Bootstrap$Modal$isFade = function (conf) {
	return A2(
		elm$core$Maybe$withDefault,
		false,
		A2(
			elm$core$Maybe$map,
			function (_n0) {
				return true;
			},
			conf.withAnimation));
};
var rundis$elm_bootstrap$Bootstrap$Modal$backdrop = F2(
	function (visibility, conf) {
		var attributes = function () {
			switch (visibility.$) {
				case 'Show':
					return _Utils_ap(
						_List_fromArray(
							[
								elm$html$Html$Attributes$classList(
								_List_fromArray(
									[
										_Utils_Tuple2('modal-backdrop', true),
										_Utils_Tuple2(
										'fade',
										rundis$elm_bootstrap$Bootstrap$Modal$isFade(conf)),
										_Utils_Tuple2('show', true)
									]))
							]),
						conf.options.hideOnBackdropClick ? _List_fromArray(
							[
								elm$html$Html$Events$onClick(
								rundis$elm_bootstrap$Bootstrap$Modal$getCloseMsg(conf))
							]) : _List_Nil);
				case 'StartClose':
					return _List_fromArray(
						[
							elm$html$Html$Attributes$classList(
							_List_fromArray(
								[
									_Utils_Tuple2('modal-backdrop', true),
									_Utils_Tuple2('fade', true),
									_Utils_Tuple2('show', true)
								]))
						]);
				case 'FadeClose':
					return _List_fromArray(
						[
							elm$html$Html$Attributes$classList(
							_List_fromArray(
								[
									_Utils_Tuple2('modal-backdrop', true),
									_Utils_Tuple2('fade', true),
									_Utils_Tuple2('show', false)
								]))
						]);
				default:
					return _List_fromArray(
						[
							elm$html$Html$Attributes$classList(
							_List_fromArray(
								[
									_Utils_Tuple2('modal-backdrop', false),
									_Utils_Tuple2(
									'fade',
									rundis$elm_bootstrap$Bootstrap$Modal$isFade(conf)),
									_Utils_Tuple2('show', false)
								]))
						]);
			}
		}();
		return _List_fromArray(
			[
				A2(elm$html$Html$div, attributes, _List_Nil)
			]);
	});
var rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$target = function (decoder) {
	return A2(elm$json$Json$Decode$field, 'target', decoder);
};
var rundis$elm_bootstrap$Bootstrap$Modal$containerClickDecoder = function (closeMsg) {
	return A2(
		elm$json$Json$Decode$andThen,
		function (c) {
			return A2(elm$core$String$contains, 'elm-bootstrap-modal', c) ? elm$json$Json$Decode$succeed(closeMsg) : elm$json$Json$Decode$fail('ignoring');
		},
		rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$target(rundis$elm_bootstrap$Bootstrap$Utilities$DomHelper$className));
};
var rundis$elm_bootstrap$Bootstrap$Modal$display = F2(
	function (visibility, conf) {
		switch (visibility.$) {
			case 'Show':
				return _List_fromArray(
					[
						A2(elm$html$Html$Attributes$style, 'pointer-events', 'none'),
						A2(elm$html$Html$Attributes$style, 'display', 'block'),
						elm$html$Html$Attributes$classList(
						_List_fromArray(
							[
								_Utils_Tuple2('modal', true),
								_Utils_Tuple2(
								'fade',
								rundis$elm_bootstrap$Bootstrap$Modal$isFade(conf)),
								_Utils_Tuple2('show', true)
							]))
					]);
			case 'StartClose':
				return _List_fromArray(
					[
						A2(elm$html$Html$Attributes$style, 'pointer-events', 'none'),
						A2(elm$html$Html$Attributes$style, 'display', 'block'),
						elm$html$Html$Attributes$classList(
						_List_fromArray(
							[
								_Utils_Tuple2('modal', true),
								_Utils_Tuple2('fade', true),
								_Utils_Tuple2('show', true)
							]))
					]);
			case 'FadeClose':
				return _List_fromArray(
					[
						A2(elm$html$Html$Attributes$style, 'pointer-events', 'none'),
						A2(elm$html$Html$Attributes$style, 'display', 'block'),
						elm$html$Html$Attributes$classList(
						_List_fromArray(
							[
								_Utils_Tuple2('modal', true),
								_Utils_Tuple2('fade', true),
								_Utils_Tuple2('show', false)
							])),
						A2(
						elm$html$Html$Events$on,
						'transitionend',
						elm$json$Json$Decode$succeed(conf.closeMsg))
					]);
			default:
				return _List_fromArray(
					[
						A2(elm$html$Html$Attributes$style, 'height', '0px'),
						A2(elm$html$Html$Attributes$style, 'display', 'block'),
						elm$html$Html$Attributes$classList(
						_List_fromArray(
							[
								_Utils_Tuple2('modal', true),
								_Utils_Tuple2(
								'fade',
								rundis$elm_bootstrap$Bootstrap$Modal$isFade(conf)),
								_Utils_Tuple2('show', false)
							]))
					]);
		}
	});
var rundis$elm_bootstrap$Bootstrap$Modal$modalClass = function (size) {
	var _n0 = rundis$elm_bootstrap$Bootstrap$General$Internal$screenSizeOption(size);
	if (_n0.$ === 'Just') {
		var s = _n0.a;
		return _List_fromArray(
			[
				elm$html$Html$Attributes$class('modal-' + s)
			]);
	} else {
		return _List_Nil;
	}
};
var rundis$elm_bootstrap$Bootstrap$Modal$modalAttributes = function (options) {
	return _Utils_ap(
		_List_fromArray(
			[
				elm$html$Html$Attributes$classList(
				_List_fromArray(
					[
						_Utils_Tuple2('modal-dialog', true),
						_Utils_Tuple2('modal-dialog-centered', options.centered)
					])),
				A2(elm$html$Html$Attributes$style, 'pointer-events', 'auto')
			]),
		A2(
			elm$core$Maybe$withDefault,
			_List_Nil,
			A2(elm$core$Maybe$map, rundis$elm_bootstrap$Bootstrap$Modal$modalClass, options.modalSize)));
};
var rundis$elm_bootstrap$Bootstrap$Modal$renderBody = function (maybeBody) {
	if (maybeBody.$ === 'Just') {
		var cfg = maybeBody.a.a;
		return elm$core$Maybe$Just(
			A2(
				elm$html$Html$div,
				A2(
					elm$core$List$cons,
					elm$html$Html$Attributes$class('modal-body'),
					cfg.attributes),
				cfg.children));
	} else {
		return elm$core$Maybe$Nothing;
	}
};
var rundis$elm_bootstrap$Bootstrap$Modal$renderFooter = function (maybeFooter) {
	if (maybeFooter.$ === 'Just') {
		var cfg = maybeFooter.a.a;
		return elm$core$Maybe$Just(
			A2(
				elm$html$Html$div,
				A2(
					elm$core$List$cons,
					elm$html$Html$Attributes$class('modal-footer'),
					cfg.attributes),
				cfg.children));
	} else {
		return elm$core$Maybe$Nothing;
	}
};
var rundis$elm_bootstrap$Bootstrap$Modal$closeButton = function (closeMsg) {
	return A2(
		elm$html$Html$button,
		_List_fromArray(
			[
				elm$html$Html$Attributes$class('close'),
				elm$html$Html$Events$onClick(closeMsg)
			]),
		_List_fromArray(
			[
				elm$html$Html$text('×')
			]));
};
var rundis$elm_bootstrap$Bootstrap$Modal$renderHeader = function (conf_) {
	var _n0 = conf_.header;
	if (_n0.$ === 'Just') {
		var cfg = _n0.a.a;
		return elm$core$Maybe$Just(
			A2(
				elm$html$Html$div,
				A2(
					elm$core$List$cons,
					elm$html$Html$Attributes$class('modal-header'),
					cfg.attributes),
				_Utils_ap(
					cfg.children,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Modal$closeButton(
							rundis$elm_bootstrap$Bootstrap$Modal$getCloseMsg(conf_))
						]))));
	} else {
		return elm$core$Maybe$Nothing;
	}
};
var rundis$elm_bootstrap$Bootstrap$Modal$view = F2(
	function (visibility, _n0) {
		var conf = _n0.a;
		return A2(
			elm$html$Html$div,
			_List_Nil,
			_Utils_ap(
				_List_fromArray(
					[
						A2(
						elm$html$Html$div,
						_Utils_ap(
							_List_fromArray(
								[
									elm$html$Html$Attributes$tabindex(-1)
								]),
							A2(rundis$elm_bootstrap$Bootstrap$Modal$display, visibility, conf)),
						_List_fromArray(
							[
								A2(
								elm$html$Html$div,
								_Utils_ap(
									_List_fromArray(
										[
											A2(elm$html$Html$Attributes$attribute, 'role', 'document'),
											elm$html$Html$Attributes$class('elm-bootstrap-modal')
										]),
									_Utils_ap(
										rundis$elm_bootstrap$Bootstrap$Modal$modalAttributes(conf.options),
										conf.options.hideOnBackdropClick ? _List_fromArray(
											[
												A2(
												elm$html$Html$Events$on,
												'click',
												rundis$elm_bootstrap$Bootstrap$Modal$containerClickDecoder(conf.closeMsg))
											]) : _List_Nil)),
								_List_fromArray(
									[
										A2(
										elm$html$Html$div,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('modal-content')
											]),
										A2(
											elm$core$List$filterMap,
											elm$core$Basics$identity,
											_List_fromArray(
												[
													rundis$elm_bootstrap$Bootstrap$Modal$renderHeader(conf),
													rundis$elm_bootstrap$Bootstrap$Modal$renderBody(conf.body),
													rundis$elm_bootstrap$Bootstrap$Modal$renderFooter(conf.footer)
												])))
									]))
							]))
					]),
				A2(rundis$elm_bootstrap$Bootstrap$Modal$backdrop, visibility, conf)));
	});
var rundis$elm_bootstrap$Bootstrap$Modal$withAnimation = F2(
	function (animateMsg, _n0) {
		var conf = _n0.a;
		return rundis$elm_bootstrap$Bootstrap$Modal$Config(
			_Utils_update(
				conf,
				{
					withAnimation: elm$core$Maybe$Just(animateMsg)
				}));
	});
var author$project$Modals$History$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Modal$view,
		model.modal,
		A3(
			rundis$elm_bootstrap$Bootstrap$Modal$footer,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(
									author$project$Modals$History$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Close')
						]))
				]),
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$body,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('scrollable-modal-row')
											]))
									]),
								author$project$Modals$History$viewHistory(model))
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$header,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('modal-title modal-header-success')
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text('History')
								]))
						]),
					A2(
						rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
						author$project$Modals$History$AnimateModal,
						rundis$elm_bootstrap$Bootstrap$Modal$large(
							rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$History$ModalClose)))))));
};
var author$project$Modals$Mkdir$CreateDir = function (a) {
	return {$: 'CreateDir', a: a};
};
var author$project$Modals$Mkdir$ModalClose = {$: 'ModalClose'};
var author$project$Modals$Mkdir$showPathCollision = F2(
	function (model, doesExist) {
		return doesExist ? A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('text-left')
				]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('fas fa-md fa-exclamation-triangle text-warning')
						]),
					_List_Nil),
					A2(
					elm$html$Html$span,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('text-muted')
						]),
					_List_fromArray(
						[
							elm$html$Html$text(' »' + (model.inputName + '« exists already. Please choose another name.\u00a0\u00a0\u00a0'))
						]))
				])) : A2(elm$html$Html$span, _List_Nil, _List_Nil);
	});
var author$project$Modals$Mkdir$InputChanged = function (a) {
	return {$: 'InputChanged', a: a};
};
var elm$html$Html$Attributes$autofocus = elm$html$Html$Attributes$boolProperty('autofocus');
var author$project$Modals$Mkdir$viewMkdirContent = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Form$Input$text(
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Form$Input$id('mkdir-input'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$large,
							rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Directory name'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$onInput(author$project$Modals$Mkdir$InputChanged),
							rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$autofocus(true)
								]))
						])),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					function () {
					var _n0 = model.state;
					if (_n0.$ === 'Ready') {
						return elm$html$Html$text('');
					} else {
						var message = _n0.a;
						return A5(author$project$Util$buildAlert, model.alert, author$project$Modals$Mkdir$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$danger, 'Oh no!', 'Could not create directory: ' + message);
					}
				}()
				]))
		]);
};
var author$project$Modals$Mkdir$view = F3(
	function (model, url, existChecker) {
		var path = author$project$Util$urlToPath(url);
		var hasPathCollision = existChecker(model.inputName);
		return A2(
			rundis$elm_bootstrap$Bootstrap$Modal$view,
			model.modal,
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$footer,
				_List_Nil,
				_List_fromArray(
					[
						A2(author$project$Modals$Mkdir$showPathCollision, model, hasPathCollision),
						A2(
						rundis$elm_bootstrap$Bootstrap$Button$button,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Button$primary,
								rundis$elm_bootstrap$Bootstrap$Button$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onClick(
										author$project$Modals$Mkdir$CreateDir(
											A2(author$project$Modals$Mkdir$pathFromUrl, url, model))),
										elm$html$Html$Attributes$type_('submit'),
										elm$html$Html$Attributes$disabled(
										(!elm$core$String$length(model.inputName)) || (function () {
											var _n0 = model.state;
											if (_n0.$ === 'Fail') {
												return true;
											} else {
												return false;
											}
										}() || hasPathCollision))
									]))
							]),
						_List_fromArray(
							[
								elm$html$Html$text('Create')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Button$button,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
								rundis$elm_bootstrap$Bootstrap$Button$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onClick(
										author$project$Modals$Mkdir$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
									]))
							]),
						_List_fromArray(
							[
								elm$html$Html$text('Cancel')
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$body,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Grid$row,
									_List_Nil,
									author$project$Modals$Mkdir$viewMkdirContent(model))
								]))
						]),
					A3(
						rundis$elm_bootstrap$Bootstrap$Modal$header,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('modal-title modal-header-primary')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$h4,
								_List_Nil,
								_List_fromArray(
									[
										elm$html$Html$text('Create a new directory in '),
										A2(
										elm$html$Html$span,
										_List_Nil,
										_List_fromArray(
											[
												elm$html$Html$text('»'),
												elm$html$Html$text(
												(path === '/') ? 'Home' : path),
												elm$html$Html$text('«')
											]))
									]))
							]),
						A2(
							rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
							author$project$Modals$Mkdir$AnimateModal,
							rundis$elm_bootstrap$Bootstrap$Modal$large(
								rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$Mkdir$ModalClose)))))));
	});
var author$project$Modals$MoveCopy$DoAction = {$: 'DoAction'};
var author$project$Modals$MoveCopy$ModalClose = {$: 'ModalClose'};
var author$project$Modals$MoveCopy$typeToString = function (typ) {
	if (typ.$ === 'Move') {
		return 'Move';
	} else {
		return 'Copy';
	}
};
var author$project$Modals$MoveCopy$DirChosen = function (a) {
	return {$: 'DirChosen', a: a};
};
var author$project$Modals$MoveCopy$SearchInput = function (a) {
	return {$: 'SearchInput', a: a};
};
var author$project$Modals$MoveCopy$filterAllDirs = F2(
	function (filter, dirs) {
		var lowerFilter = elm$core$String$toLower(filter);
		return A2(
			elm$core$List$filter,
			elm$core$String$contains(lowerFilter),
			dirs);
	});
var author$project$Modals$MoveCopy$viewDirEntry = F2(
	function (clickMsg, path) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Table$tr,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Events$onClick(
								clickMsg(path)))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-lg fa-folder text-xs-right file-list-icon')
								]),
							_List_Nil)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
							elm$html$Html$Events$onClick(
								clickMsg(path)))
						]),
					_List_fromArray(
						[
							elm$html$Html$text(path)
						]))
				]));
	});
var rundis$elm_bootstrap$Bootstrap$Table$HeadAttr = function (a) {
	return {$: 'HeadAttr', a: a};
};
var rundis$elm_bootstrap$Bootstrap$Table$headAttr = function (attr_) {
	return rundis$elm_bootstrap$Bootstrap$Table$HeadAttr(attr_);
};
var author$project$Modals$MoveCopy$viewDirList = F3(
	function (clickMsg, filter, dirs) {
		return rundis$elm_bootstrap$Bootstrap$Table$table(
			{
				options: _List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Table$hover]),
				tbody: A2(
					rundis$elm_bootstrap$Bootstrap$Table$tbody,
					_List_Nil,
					A2(
						elm$core$List$map,
						author$project$Modals$MoveCopy$viewDirEntry(clickMsg),
						A2(author$project$Modals$MoveCopy$filterAllDirs, filter, dirs))),
				thead: A2(
					rundis$elm_bootstrap$Bootstrap$Table$thead,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Table$headAttr(
							A2(elm$html$Html$Attributes$style, 'display', 'none'))
						]),
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$tr,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '10%'))
										]),
									_List_Nil),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '90%'))
										]),
									_List_Nil)
								]))
						]))
			});
	});
var author$project$Modals$MoveCopy$viewSearchBox = F2(
	function (searchMsg, filter) {
		return rundis$elm_bootstrap$Bootstrap$Form$InputGroup$view(
			A2(
				rundis$elm_bootstrap$Bootstrap$Form$InputGroup$attrs,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('stylish-input-group input-group')
					]),
				A2(
					rundis$elm_bootstrap$Bootstrap$Form$InputGroup$successors,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Form$InputGroup$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('input-group-addon')
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$button,
									_List_Nil,
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('fas fa-search fa-xs input-group-addon')
												]),
											_List_Nil)
										]))
								]))
						]),
					rundis$elm_bootstrap$Bootstrap$Form$InputGroup$config(
						rundis$elm_bootstrap$Bootstrap$Form$InputGroup$text(
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Filter directory list'),
									rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
									_List_fromArray(
										[
											elm$html$Html$Events$onInput(searchMsg),
											elm$html$Html$Attributes$value(filter)
										]))
								]))))));
	});
var author$project$Modals$MoveCopy$viewContent = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					function () {
					var _n0 = model.state;
					switch (_n0.$) {
						case 'Ready':
							var dirs = _n0.a;
							return A2(
								elm$html$Html$div,
								_List_Nil,
								_List_fromArray(
									[
										A2(author$project$Modals$MoveCopy$viewSearchBox, author$project$Modals$MoveCopy$SearchInput, model.filter),
										A3(author$project$Modals$MoveCopy$viewDirList, author$project$Modals$MoveCopy$DirChosen, model.filter, dirs)
									]));
						case 'Loading':
							return elm$html$Html$text('Loading.');
						default:
							var message = _n0.a;
							return A5(author$project$Util$buildAlert, model.alert, author$project$Modals$MoveCopy$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$danger, 'Oh no!', 'Could not move or copy path: ' + message);
					}
				}()
				]))
		]);
};
var author$project$Modals$MoveCopy$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Modal$view,
		model.modal,
		A3(
			rundis$elm_bootstrap$Bootstrap$Modal$footer,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$primary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(author$project$Modals$MoveCopy$DoAction),
									elm$html$Html$Attributes$type_('submit'),
									elm$html$Html$Attributes$disabled(
									(!elm$core$String$length(model.destPath)) || function () {
										var _n0 = model.state;
										if (_n0.$ === 'Fail') {
											return true;
										} else {
											return false;
										}
									}())
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text(
							author$project$Modals$MoveCopy$typeToString(model.action))
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(
									author$project$Modals$MoveCopy$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Cancel')
						]))
				]),
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$body,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('scrollable-modal-row')
											]))
									]),
								author$project$Modals$MoveCopy$viewContent(model))
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$header,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('modal-title modal-header-primary')
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text(
									author$project$Modals$MoveCopy$typeToString(model.action) + ' '),
									A2(
									elm$html$Html$span,
									_List_Nil,
									_List_fromArray(
										[
											elm$html$Html$text('»'),
											elm$html$Html$text(
											author$project$Util$basename(model.sourcePath)),
											elm$html$Html$text('«')
										])),
									(elm$core$String$length(model.destPath) > 0) ? A2(
									elm$html$Html$span,
									_List_Nil,
									_List_fromArray(
										[
											elm$html$Html$text(' into »'),
											elm$html$Html$text(model.destPath),
											elm$html$Html$text('«')
										])) : elm$html$Html$text(' into ...')
								]))
						]),
					A2(
						rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
						author$project$Modals$MoveCopy$AnimateModal,
						rundis$elm_bootstrap$Bootstrap$Modal$large(
							rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$MoveCopy$ModalClose)))))));
};
var author$project$Modals$Remove$ModalClose = {$: 'ModalClose'};
var author$project$Modals$Remove$RemoveAll = function (a) {
	return {$: 'RemoveAll', a: a};
};
var author$project$Modals$Remove$pluralizeItems = function (count) {
	return (count === 1) ? 'item' : 'items';
};
var author$project$Modals$Remove$viewRemoveContent = F2(
	function (model, nSelected) {
		return _List_fromArray(
			[
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$col,
				_List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
				_List_fromArray(
					[
						function () {
						var _n0 = model.state;
						if (_n0.$ === 'Ready') {
							return elm$html$Html$text(
								'This would remove the ' + (elm$core$String$fromInt(nSelected) + (' selected ' + (author$project$Modals$Remove$pluralizeItems(nSelected) + '.'))));
						} else {
							var message = _n0.a;
							return A5(author$project$Util$buildAlert, model.alert, author$project$Modals$Remove$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$danger, 'Oh no!', 'Could not remove directory: ' + message);
						}
					}()
					]))
			]);
	});
var rundis$elm_bootstrap$Bootstrap$Internal$Button$Warning = {$: 'Warning'};
var rundis$elm_bootstrap$Bootstrap$Button$warning = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Roled(rundis$elm_bootstrap$Bootstrap$Internal$Button$Warning));
var author$project$Modals$Remove$view = F2(
	function (model, selectedPaths) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Modal$view,
			model.modal,
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$footer,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Button$button,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Button$warning,
								rundis$elm_bootstrap$Bootstrap$Button$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onClick(
										author$project$Modals$Remove$RemoveAll(selectedPaths)),
										elm$html$Html$Attributes$disabled(
										function () {
											var _n0 = model.state;
											if (_n0.$ === 'Fail') {
												return true;
											} else {
												return false;
											}
										}())
									]))
							]),
						_List_fromArray(
							[
								elm$html$Html$text('Remove')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Button$button,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
								rundis$elm_bootstrap$Bootstrap$Button$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onClick(
										author$project$Modals$Remove$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
									]))
							]),
						_List_fromArray(
							[
								elm$html$Html$text('Cancel')
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$body,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Grid$row,
									_List_Nil,
									A2(
										author$project$Modals$Remove$viewRemoveContent,
										model,
										elm$core$List$length(selectedPaths)))
								]))
						]),
					A3(
						rundis$elm_bootstrap$Bootstrap$Modal$header,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('modal-title modal-header-warning')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$h4,
								_List_Nil,
								_List_fromArray(
									[
										elm$html$Html$text('Really remove?')
									]))
							]),
						A2(
							rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
							author$project$Modals$Remove$AnimateModal,
							rundis$elm_bootstrap$Bootstrap$Modal$large(
								rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$Remove$ModalClose)))))));
	});
var author$project$Modals$Rename$DoRename = {$: 'DoRename'};
var author$project$Modals$Rename$ModalClose = {$: 'ModalClose'};
var author$project$Modals$Rename$InputChanged = function (a) {
	return {$: 'InputChanged', a: a};
};
var author$project$Modals$Rename$viewRenameContent = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Form$Input$text(
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Form$Input$id('rename-input'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$large,
							rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('New name'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$onInput(author$project$Modals$Rename$InputChanged),
							rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$autofocus(true)
								]))
						])),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					function () {
					var _n0 = model.state;
					if (_n0.$ === 'Ready') {
						return elm$html$Html$text('');
					} else {
						var message = _n0.a;
						return A5(author$project$Util$buildAlert, model.alert, author$project$Modals$Rename$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$danger, 'Oh no!', 'Could not rename path: ' + message);
					}
				}()
				]))
		]);
};
var author$project$Modals$Rename$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Modal$view,
		model.modal,
		A3(
			rundis$elm_bootstrap$Bootstrap$Modal$footer,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$primary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(author$project$Modals$Rename$DoRename),
									elm$html$Html$Attributes$type_('submit'),
									elm$html$Html$Attributes$disabled(
									(!elm$core$String$length(model.inputName)) || function () {
										var _n0 = model.state;
										if (_n0.$ === 'Fail') {
											return true;
										} else {
											return false;
										}
									}())
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Rename')
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(
									author$project$Modals$Rename$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Cancel')
						]))
				]),
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$body,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_Nil,
								author$project$Modals$Rename$viewRenameContent(model))
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$header,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('modal-title modal-header-primary')
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text('Rename '),
									A2(
									elm$html$Html$span,
									_List_Nil,
									_List_fromArray(
										[
											elm$html$Html$text('»'),
											elm$html$Html$text(
											author$project$Util$basename(model.currPath)),
											elm$html$Html$text('«')
										])),
									(elm$core$String$length(model.inputName) > 0) ? A2(
									elm$html$Html$span,
									_List_Nil,
									_List_fromArray(
										[
											elm$html$Html$text(' to '),
											A2(
											elm$html$Html$span,
											_List_Nil,
											_List_fromArray(
												[
													elm$html$Html$text('»'),
													elm$html$Html$text(model.inputName),
													elm$html$Html$text('«')
												]))
										])) : elm$html$Html$text('')
								]))
						]),
					A2(
						rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
						author$project$Modals$Rename$AnimateModal,
						rundis$elm_bootstrap$Bootstrap$Modal$large(
							rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$Rename$ModalClose)))))));
};
var author$project$Modals$Share$ModalClose = {$: 'ModalClose'};
var author$project$Modals$Share$formatEntry = F2(
	function (url, path) {
		var link = author$project$Util$urlPrefixToString(url) + ('get' + author$project$Util$urlEncodePath(path));
		return A2(
			elm$html$Html$li,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					elm$html$Html$a,
					_List_fromArray(
						[
							elm$html$Html$Attributes$href(link)
						]),
					_List_fromArray(
						[
							elm$html$Html$text(link)
						]))
				]));
	});
var elm$html$Html$b = _VirtualDom_node('b');
var author$project$Modals$Share$viewShare = F2(
	function (model, url) {
		return _List_fromArray(
			[
				A2(
				rundis$elm_bootstrap$Bootstrap$Grid$col,
				_List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
				_List_fromArray(
					[
						A2(
						elm$html$Html$p,
						_List_Nil,
						_List_fromArray(
							[
								elm$html$Html$text('Use those links to share the selected files with people that do not use brig.')
							])),
						A2(
						elm$html$Html$p,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								elm$html$Html$b,
								_List_Nil,
								_List_fromArray(
									[
										elm$html$Html$text('Note:')
									])),
								elm$html$Html$text(' Remember, they still need to authenticate themselves.')
							])),
						A2(
						elm$html$Html$ul,
						_List_fromArray(
							[
								elm$html$Html$Attributes$id('share-list')
							]),
						A2(
							elm$core$List$map,
							author$project$Modals$Share$formatEntry(url),
							model.paths))
					]))
			]);
	});
var author$project$Modals$Share$view = F2(
	function (model, url) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Modal$view,
			model.modal,
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$footer,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Button$button,
						_List_fromArray(
							[
								rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
								rundis$elm_bootstrap$Bootstrap$Button$attrs(
								_List_fromArray(
									[
										elm$html$Html$Events$onClick(
										author$project$Modals$Share$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
									]))
							]),
						_List_fromArray(
							[
								elm$html$Html$text('Close')
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$body,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Grid$row,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('scrollable-modal-row')
												]))
										]),
									A2(author$project$Modals$Share$viewShare, model, url))
								]))
						]),
					A3(
						rundis$elm_bootstrap$Bootstrap$Modal$header,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('modal-title modal-header-primary')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$h4,
								_List_Nil,
								_List_fromArray(
									[
										elm$html$Html$text('Share hyperlinks')
									]))
							]),
						A2(
							rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
							author$project$Modals$Share$AnimateModal,
							rundis$elm_bootstrap$Bootstrap$Modal$large(
								rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$Share$ModalClose)))))));
	});
var author$project$Routes$Ls$existsInCurr = F2(
	function (model, name) {
		var _n0 = model.state;
		if (_n0.$ === 'Success') {
			var actualModel = _n0.a;
			var _n1 = actualModel.isFiltered;
			if (_n1) {
				return false;
			} else {
				return A2(
					elm$core$List$any,
					function (e) {
						return _Utils_eq(
							name,
							author$project$Util$basename(e.path));
					},
					actualModel.entries);
			}
		} else {
			return false;
		}
	});
var author$project$Routes$Ls$buildModals = function (model) {
	var paths = author$project$Routes$Ls$selectedPaths(model);
	return A2(
		elm$html$Html$span,
		_List_Nil,
		_List_fromArray(
			[
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$HistoryMsg,
				author$project$Modals$History$view(model.historyState)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$RenameMsg,
				author$project$Modals$Rename$view(model.renameState)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$MoveMsg,
				author$project$Modals$MoveCopy$view(model.moveState)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$CopyMsg,
				author$project$Modals$MoveCopy$view(model.copyState)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$MkdirMsg,
				A3(
					author$project$Modals$Mkdir$view,
					model.mkdirState,
					model.url,
					author$project$Routes$Ls$existsInCurr(model))),
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$RemoveMsg,
				A2(author$project$Modals$Remove$view, model.removeState, paths)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Ls$ShareMsg,
				A2(author$project$Modals$Share$view, model.shareState, model.url))
			]));
};
var author$project$Modals$RemoteAdd$ModalClose = {$: 'ModalClose'};
var author$project$Modals$RemoteAdd$RemoteAdd = {$: 'RemoteAdd'};
var author$project$Modals$RemoteAdd$AcceptPushChanged = function (a) {
	return {$: 'AcceptPushChanged', a: a};
};
var author$project$Modals$RemoteAdd$AutoUpdateChanged = function (a) {
	return {$: 'AutoUpdateChanged', a: a};
};
var author$project$Modals$RemoteAdd$FingerprintInputChanged = function (a) {
	return {$: 'FingerprintInputChanged', a: a};
};
var author$project$Modals$RemoteAdd$NameInputChanged = function (a) {
	return {$: 'NameInputChanged', a: a};
};
var author$project$Modals$RemoteAdd$ConflictStrategyChanged = function (a) {
	return {$: 'ConflictStrategyChanged', a: a};
};
var author$project$Modals$RemoteAdd$showCurrentConflictStrategy = function (model) {
	var _n0 = model.conflictStrategy;
	switch (_n0) {
		case '':
			return A2(
				elm$html$Html$span,
				_List_Nil,
				_List_fromArray(
					[
						elm$html$Html$text('Marker '),
						A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('fas fa-marker')
							]),
						_List_Nil)
					]));
		case 'ignore':
			return A2(
				elm$html$Html$span,
				_List_Nil,
				_List_fromArray(
					[
						elm$html$Html$text('Ignore '),
						A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('fas fa-eject')
							]),
						_List_Nil)
					]));
		case 'marker':
			return A2(
				elm$html$Html$span,
				_List_Nil,
				_List_fromArray(
					[
						elm$html$Html$text('Marker '),
						A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('fas fa-marker')
							]),
						_List_Nil)
					]));
		case 'embrace':
			return A2(
				elm$html$Html$span,
				_List_Nil,
				_List_fromArray(
					[
						elm$html$Html$text('Embrace '),
						A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('fas fa-handshake')
							]),
						_List_Nil)
					]));
		default:
			return A2(
				elm$html$Html$span,
				_List_Nil,
				_List_fromArray(
					[
						elm$html$Html$text('Unknown '),
						A2(
						elm$html$Html$span,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('fas fa-question')
							]),
						_List_Nil)
					]));
	}
};
var author$project$Modals$RemoteAdd$viewConflictDropdown = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Dropdown$dropdown,
		model.conflictDropdown,
		{
			items: _List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
					_List_fromArray(
						[
							elm$html$Html$Events$onClick(
							author$project$Modals$RemoteAdd$ConflictStrategyChanged('ignore'))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-md fa-eject')
								]),
							_List_Nil),
							elm$html$Html$text(' Ignore')
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
					_List_fromArray(
						[
							elm$html$Html$Events$onClick(
							author$project$Modals$RemoteAdd$ConflictStrategyChanged('marker'))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-md fa-marker')
								]),
							_List_Nil),
							elm$html$Html$text(' Marker')
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
					_List_fromArray(
						[
							elm$html$Html$Events$onClick(
							author$project$Modals$RemoteAdd$ConflictStrategyChanged('embrace'))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-md fa-handshake')
								]),
							_List_Nil),
							elm$html$Html$text(' Embrace')
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
					_List_fromArray(
						[
							elm$html$Html$Events$onClick(
							author$project$Modals$RemoteAdd$ConflictStrategyChanged(''))
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-md fa-eraser')
								]),
							_List_Nil),
							elm$html$Html$text(' Default')
						]))
				]),
			options: _List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Dropdown$alignMenuRight,
					rundis$elm_bootstrap$Bootstrap$Dropdown$attrs(
					_List_fromArray(
						[
							elm$html$Html$Attributes$id('remote-add-conflict-dropdown')
						]))
				]),
			toggleButton: A2(
				rundis$elm_bootstrap$Bootstrap$Dropdown$toggle,
				_List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Button$roleLink]),
				_List_fromArray(
					[
						author$project$Modals$RemoteAdd$showCurrentConflictStrategy(model)
					])),
			toggleMsg: author$project$Modals$RemoteAdd$ConflictDropdownMsg
		});
};
var author$project$Modals$RemoteAdd$viewRemoteAddContent = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					rundis$elm_bootstrap$Bootstrap$Form$Input$text(
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Form$Input$id('remote-name-input'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$large,
							rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Remote name'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$onInput(author$project$Modals$RemoteAdd$NameInputChanged),
							rundis$elm_bootstrap$Bootstrap$Form$Input$attrs(
							_List_fromArray(
								[
									elm$html$Html$Attributes$autofocus(true)
								]))
						])),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					rundis$elm_bootstrap$Bootstrap$Form$Input$text(
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Form$Input$id('remote-fingerprint-input'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$large,
							rundis$elm_bootstrap$Bootstrap$Form$Input$placeholder('Remote fingerprint'),
							rundis$elm_bootstrap$Bootstrap$Form$Input$onInput(author$project$Modals$RemoteAdd$FingerprintInputChanged)
						])),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(
					elm$html$Html$span,
					_List_Nil,
					_List_fromArray(
						[
							A4(author$project$Util$viewToggleSwitch, author$project$Modals$RemoteAdd$AutoUpdateChanged, 'Accept automatic updates?', model.doAutoUdate, false)
						])),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(
					elm$html$Html$span,
					_List_Nil,
					_List_fromArray(
						[
							A4(author$project$Util$viewToggleSwitch, author$project$Modals$RemoteAdd$AcceptPushChanged, 'Accept other remotes pushing data to us?', model.acceptPush, false)
						])),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(
					elm$html$Html$span,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted')
								]),
							_List_fromArray(
								[
									elm$html$Html$text('The current conflict strategy is')
								])),
							author$project$Modals$RemoteAdd$viewConflictDropdown(model),
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted')
								]),
							_List_fromArray(
								[
									elm$html$Html$text('.')
								]))
						])),
					function () {
					var _n0 = model.state;
					if (_n0.$ === 'Ready') {
						return elm$html$Html$text('');
					} else {
						var message = _n0.a;
						return A5(author$project$Util$buildAlert, model.alert, author$project$Modals$RemoteAdd$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$danger, 'Oh no!', 'Could not add remote: ' + message);
					}
				}()
				]))
		]);
};
var author$project$Modals$RemoteAdd$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Modal$view,
		model.modal,
		A3(
			rundis$elm_bootstrap$Bootstrap$Modal$footer,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$primary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(author$project$Modals$RemoteAdd$RemoteAdd),
									elm$html$Html$Attributes$type_('submit'),
									elm$html$Html$Attributes$disabled(
									(!elm$core$String$length(model.name)) || ((!elm$core$String$length(model.fingerprint)) || function () {
										var _n0 = model.state;
										if (_n0.$ === 'Fail') {
											return true;
										} else {
											return false;
										}
									}()))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Create')
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(
									author$project$Modals$RemoteAdd$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Cancel')
						]))
				]),
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$body,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_Nil,
								author$project$Modals$RemoteAdd$viewRemoteAddContent(model))
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$header,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('modal-title modal-header-primary')
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text('Add a new remote')
								]))
						]),
					A2(
						rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
						author$project$Modals$RemoteAdd$AnimateModal,
						rundis$elm_bootstrap$Bootstrap$Modal$large(
							rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$RemoteAdd$ModalClose)))))));
};
var author$project$Modals$RemoteFolders$ModalClose = {$: 'ModalClose'};
var author$project$Modals$RemoteFolders$DirChosen = function (a) {
	return {$: 'DirChosen', a: a};
};
var author$project$Modals$RemoteFolders$SearchInput = function (a) {
	return {$: 'SearchInput', a: a};
};
var author$project$Modals$RemoteFolders$FolderRemove = function (a) {
	return {$: 'FolderRemove', a: a};
};
var author$project$Modals$RemoteFolders$ReadOnlyChanged = F2(
	function (a, b) {
		return {$: 'ReadOnlyChanged', a: a, b: b};
	});
var author$project$Modals$RemoteFolders$ConflictStrategyToggled = F2(
	function (a, b) {
		return {$: 'ConflictStrategyToggled', a: a, b: b};
	});
var author$project$Modals$RemoteFolders$conflictStrategyToIconName = function (strategy) {
	switch (strategy) {
		case '':
			return 'fa-marker text-muted';
		case 'ignore':
			return 'fa-eject';
		case 'marker':
			return 'fa-marker';
		case 'embrace':
			return 'fa-handshake';
		default:
			return 'fa-question';
	}
};
var author$project$Modals$RemoteFolders$viewConflictDropdown = F2(
	function (model, folder) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Dropdown$dropdown,
			A2(
				elm$core$Maybe$withDefault,
				rundis$elm_bootstrap$Bootstrap$Dropdown$initialState,
				A2(elm$core$Dict$get, folder.folder, model.conflictDropdowns)),
			{
				items: _List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Modals$RemoteFolders$ConflictStrategyToggled, folder.folder, 'ignore'))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-eject')
									]),
								_List_Nil),
								elm$html$Html$text(' Ignore')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Modals$RemoteFolders$ConflictStrategyToggled, folder.folder, 'marker'))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-marker')
									]),
								_List_Nil),
								elm$html$Html$text(' Marker')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Modals$RemoteFolders$ConflictStrategyToggled, folder.folder, 'embrace'))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-handshake')
									]),
								_List_Nil),
								elm$html$Html$text(' Embrace')
							])),
						A2(
						rundis$elm_bootstrap$Bootstrap$Dropdown$buttonItem,
						_List_fromArray(
							[
								elm$html$Html$Events$onClick(
								A2(author$project$Modals$RemoteFolders$ConflictStrategyToggled, folder.folder, ''))
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$span,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('fas fa-md fa-eraser')
									]),
								_List_Nil),
								elm$html$Html$text(' Default')
							]))
					]),
				options: _List_fromArray(
					[rundis$elm_bootstrap$Bootstrap$Dropdown$alignMenuRight]),
				toggleButton: A2(
					rundis$elm_bootstrap$Bootstrap$Dropdown$toggle,
					_List_fromArray(
						[rundis$elm_bootstrap$Bootstrap$Button$roleLink]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas'),
									elm$html$Html$Attributes$class(
									author$project$Modals$RemoteFolders$conflictStrategyToIconName(folder.conflictStrategy))
								]),
							_List_Nil)
						])),
				toggleMsg: author$project$Modals$RemoteFolders$ConflictDropdownMsg(folder.folder)
			});
	});
var author$project$Modals$RemoteFolders$viewFolder = F2(
	function (model, folder) {
		return A2(
			rundis$elm_bootstrap$Bootstrap$Table$tr,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('fas fa-md fa-folder text-muted')
								]),
							_List_Nil)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							elm$html$Html$text(folder.folder)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(author$project$Modals$RemoteFolders$viewConflictDropdown, model, folder)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A4(
							author$project$Util$viewToggleSwitch,
							author$project$Modals$RemoteFolders$ReadOnlyChanged(folder.folder),
							'',
							folder.readOnly,
							false)
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Table$td,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Button$button,
							_List_fromArray(
								[
									rundis$elm_bootstrap$Bootstrap$Button$attrs(
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('close'),
											elm$html$Html$Events$onClick(
											author$project$Modals$RemoteFolders$FolderRemove(folder.folder))
										]))
								]),
							_List_fromArray(
								[
									A2(
									elm$html$Html$span,
									_List_fromArray(
										[
											elm$html$Html$Attributes$class('fas fa-xs fa-times text-muted')
										]),
									_List_Nil)
								]))
						]))
				]));
	});
var author$project$Modals$RemoteFolders$viewFolders = F2(
	function (model, remote) {
		return rundis$elm_bootstrap$Bootstrap$Table$table(
			{
				options: _List_fromArray(
					[
						rundis$elm_bootstrap$Bootstrap$Table$hover,
						rundis$elm_bootstrap$Bootstrap$Table$attr(
						elm$html$Html$Attributes$class('borderless-table'))
					]),
				tbody: A2(
					rundis$elm_bootstrap$Bootstrap$Table$tbody,
					_List_Nil,
					A2(
						elm$core$List$map,
						function (f) {
							return A2(author$project$Modals$RemoteFolders$viewFolder, model, f);
						},
						remote.folders)),
				thead: A2(
					rundis$elm_bootstrap$Bootstrap$Table$thead,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							rundis$elm_bootstrap$Bootstrap$Table$tr,
							_List_Nil,
							_List_fromArray(
								[
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '5%'))
										]),
									_List_fromArray(
										[
											elm$html$Html$text('')
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '55%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted small')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Name')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '20%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted small')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Conflict Strategy')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '15%'))
										]),
									_List_fromArray(
										[
											A2(
											elm$html$Html$span,
											_List_fromArray(
												[
													elm$html$Html$Attributes$class('text-muted small')
												]),
											_List_fromArray(
												[
													elm$html$Html$text('Read Only?')
												]))
										])),
									A2(
									rundis$elm_bootstrap$Bootstrap$Table$th,
									_List_fromArray(
										[
											rundis$elm_bootstrap$Bootstrap$Table$cellAttr(
											A2(elm$html$Html$Attributes$style, 'width', '5%'))
										]),
									_List_Nil)
								]))
						]))
			});
	});
var elm_community$list_extra$List$Extra$uniqueHelp = F4(
	function (f, existing, remaining, accumulator) {
		uniqueHelp:
		while (true) {
			if (!remaining.b) {
				return elm$core$List$reverse(accumulator);
			} else {
				var first = remaining.a;
				var rest = remaining.b;
				var computedFirst = f(first);
				if (A2(elm$core$Set$member, computedFirst, existing)) {
					var $temp$f = f,
						$temp$existing = existing,
						$temp$remaining = rest,
						$temp$accumulator = accumulator;
					f = $temp$f;
					existing = $temp$existing;
					remaining = $temp$remaining;
					accumulator = $temp$accumulator;
					continue uniqueHelp;
				} else {
					var $temp$f = f,
						$temp$existing = A2(elm$core$Set$insert, computedFirst, existing),
						$temp$remaining = rest,
						$temp$accumulator = A2(elm$core$List$cons, first, accumulator);
					f = $temp$f;
					existing = $temp$existing;
					remaining = $temp$remaining;
					accumulator = $temp$accumulator;
					continue uniqueHelp;
				}
			}
		}
	});
var elm_community$list_extra$List$Extra$uniqueBy = F2(
	function (f, list) {
		return A4(elm_community$list_extra$List$Extra$uniqueHelp, f, elm$core$Set$empty, list, _List_Nil);
	});
var author$project$Modals$RemoteFolders$viewMaybeFolders = F2(
	function (model, remote) {
		var folders = A2(
			elm_community$list_extra$List$Extra$uniqueBy,
			function ($) {
				return $.folder;
			},
			remote.folders);
		return (elm$core$List$length(folders) <= 0) ? A2(
			elm$html$Html$span,
			_List_fromArray(
				[
					elm$html$Html$Attributes$class('text-muted text-center')
				]),
			_List_fromArray(
				[
					elm$html$Html$text('No folders. This means this user can see everthing.'),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					elm$html$Html$text('Add a new folder below to limit what this remote can see.'),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(elm$html$Html$br, _List_Nil, _List_Nil)
				])) : A2(
			elm$html$Html$div,
			_List_Nil,
			_List_fromArray(
				[
					A2(author$project$Modals$RemoteFolders$viewFolders, model, remote),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(elm$html$Html$hr, _List_Nil, _List_Nil)
				]));
	});
var author$project$Modals$RemoteFolders$viewRemoteFoldersContent = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					A2(
					elm$html$Html$h4,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted text-center')
								]),
							_List_fromArray(
								[
									elm$html$Html$text('Visible folders')
								]))
						])),
					A2(author$project$Modals$RemoteFolders$viewMaybeFolders, model, model.remote),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(elm$html$Html$br, _List_Nil, _List_Nil),
					A2(
					elm$html$Html$h4,
					_List_Nil,
					_List_fromArray(
						[
							A2(
							elm$html$Html$span,
							_List_fromArray(
								[
									elm$html$Html$Attributes$class('text-muted text-center')
								]),
							_List_fromArray(
								[
									elm$html$Html$text('All folders')
								]))
						])),
					A2(author$project$Modals$MoveCopy$viewSearchBox, author$project$Modals$RemoteFolders$SearchInput, model.filter),
					A3(author$project$Modals$MoveCopy$viewDirList, author$project$Modals$RemoteFolders$DirChosen, model.filter, model.allDirs),
					function () {
					var _n0 = model.state;
					if (_n0.$ === 'Ready') {
						return elm$html$Html$text('');
					} else {
						var message = _n0.a;
						return A5(author$project$Util$buildAlert, model.alert, author$project$Modals$RemoteFolders$AlertMsg, rundis$elm_bootstrap$Bootstrap$Alert$danger, 'Oh no!', 'Could not add remote: ' + message);
					}
				}()
				]))
		]);
};
var author$project$Modals$RemoteFolders$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Modal$view,
		model.modal,
		A3(
			rundis$elm_bootstrap$Bootstrap$Modal$footer,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(
									author$project$Modals$RemoteFolders$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Close')
						]))
				]),
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$body,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_fromArray(
									[
										rundis$elm_bootstrap$Bootstrap$Grid$Row$attrs(
										_List_fromArray(
											[
												A2(elm$html$Html$Attributes$style, 'min-width', '60vh'),
												elm$html$Html$Attributes$class('scrollable-modal-row')
											]))
									]),
								author$project$Modals$RemoteFolders$viewRemoteFoldersContent(model))
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$header,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('modal-title modal-header-primary')
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text('Edit folders of »'),
									elm$html$Html$text(model.remote.name),
									elm$html$Html$text('«')
								]))
						]),
					A2(
						rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
						author$project$Modals$RemoteFolders$AnimateModal,
						rundis$elm_bootstrap$Bootstrap$Modal$large(
							rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$RemoteFolders$ModalClose)))))));
};
var author$project$Modals$RemoteRemove$DoRemove = {$: 'DoRemove'};
var author$project$Modals$RemoteRemove$ModalClose = {$: 'ModalClose'};
var author$project$Modals$RemoteRemove$viewRemoteAddContent = function (model) {
	return _List_fromArray(
		[
			A2(
			rundis$elm_bootstrap$Bootstrap$Grid$col,
			_List_fromArray(
				[rundis$elm_bootstrap$Bootstrap$Grid$Col$xs12]),
			_List_fromArray(
				[
					elm$html$Html$text('Removing »' + (model.name + ('« cannot be reverted. If you are the last one caching the data of this remote,' + ' the data might vanish forever and cannot be restored.')))
				]))
		]);
};
var rundis$elm_bootstrap$Bootstrap$Button$danger = rundis$elm_bootstrap$Bootstrap$Internal$Button$Coloring(
	rundis$elm_bootstrap$Bootstrap$Internal$Button$Roled(rundis$elm_bootstrap$Bootstrap$Internal$Button$Danger));
var author$project$Modals$RemoteRemove$view = function (model) {
	return A2(
		rundis$elm_bootstrap$Bootstrap$Modal$view,
		model.modal,
		A3(
			rundis$elm_bootstrap$Bootstrap$Modal$footer,
			_List_Nil,
			_List_fromArray(
				[
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$danger,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(author$project$Modals$RemoteRemove$DoRemove),
									elm$html$Html$Attributes$type_('submit')
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Remove')
						])),
					A2(
					rundis$elm_bootstrap$Bootstrap$Button$button,
					_List_fromArray(
						[
							rundis$elm_bootstrap$Bootstrap$Button$outlinePrimary,
							rundis$elm_bootstrap$Bootstrap$Button$attrs(
							_List_fromArray(
								[
									elm$html$Html$Events$onClick(
									author$project$Modals$RemoteRemove$AnimateModal(rundis$elm_bootstrap$Bootstrap$Modal$hiddenAnimated))
								]))
						]),
					_List_fromArray(
						[
							elm$html$Html$text('Cancel')
						]))
				]),
			A3(
				rundis$elm_bootstrap$Bootstrap$Modal$body,
				_List_Nil,
				_List_fromArray(
					[
						A2(
						rundis$elm_bootstrap$Bootstrap$Grid$containerFluid,
						_List_Nil,
						_List_fromArray(
							[
								A2(
								rundis$elm_bootstrap$Bootstrap$Grid$row,
								_List_Nil,
								author$project$Modals$RemoteRemove$viewRemoteAddContent(model))
							]))
					]),
				A3(
					rundis$elm_bootstrap$Bootstrap$Modal$header,
					_List_fromArray(
						[
							elm$html$Html$Attributes$class('modal-title modal-header-danger')
						]),
					_List_fromArray(
						[
							A2(
							elm$html$Html$h4,
							_List_Nil,
							_List_fromArray(
								[
									elm$html$Html$text('Really remove?')
								]))
						]),
					A2(
						rundis$elm_bootstrap$Bootstrap$Modal$withAnimation,
						author$project$Modals$RemoteRemove$AnimateModal,
						rundis$elm_bootstrap$Bootstrap$Modal$large(
							rundis$elm_bootstrap$Bootstrap$Modal$config(author$project$Modals$RemoteRemove$ModalClose)))))));
};
var author$project$Routes$Remotes$buildModals = function (model) {
	return A2(
		elm$html$Html$span,
		_List_Nil,
		_List_fromArray(
			[
				A2(
				elm$html$Html$map,
				author$project$Routes$Remotes$RemoteAddMsg,
				author$project$Modals$RemoteAdd$view(model.remoteAddState)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Remotes$RemoteRemoveMsg,
				author$project$Modals$RemoteRemove$view(model.remoteRemoveState)),
				A2(
				elm$html$Html$map,
				author$project$Routes$Remotes$RemoteFolderMsg,
				author$project$Modals$RemoteFolders$view(model.remoteFoldersState))
			]));
};
var elm$html$Html$aside = _VirtualDom_node('aside');
var elm$html$Html$main_ = _VirtualDom_node('main');
var author$project$Main$viewMainContent = F2(
	function (model, viewState) {
		return _List_fromArray(
			[
				A2(
				elm$html$Html$div,
				_List_fromArray(
					[
						elm$html$Html$Attributes$class('container-fluid')
					]),
				_List_fromArray(
					[
						A2(
						elm$html$Html$div,
						_List_fromArray(
							[
								elm$html$Html$Attributes$class('row wrapper')
							]),
						_List_fromArray(
							[
								A2(
								elm$html$Html$aside,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('col-12 col-md-2 p-0 bg-light tabbar')
									]),
								_List_fromArray(
									[
										A2(
										elm$html$Html$nav,
										_List_fromArray(
											[
												elm$html$Html$Attributes$class('navbar navbar-expand-md navbar-light bg-align-items-start flex-md-column flex-row')
											]),
										_List_fromArray(
											[
												author$project$Main$viewAppIcon(model),
												A2(
												elm$html$Html$a,
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('navbar-toggler'),
														A2(elm$html$Html$Attributes$attribute, 'data-toggle', 'collapse'),
														A2(elm$html$Html$Attributes$attribute, 'data-target', '.sidebar')
													]),
												_List_fromArray(
													[
														A2(
														elm$html$Html$span,
														_List_fromArray(
															[
																elm$html$Html$Attributes$class('navbar-toggler-icon')
															]),
														_List_Nil)
													])),
												A2(
												elm$html$Html$div,
												_List_fromArray(
													[
														elm$html$Html$Attributes$class('collapse navbar-collapse sidebar')
													]),
												_List_fromArray(
													[
														A2(author$project$Main$viewSidebarItems, model, viewState)
													]))
											])),
										author$project$Main$viewSidebarBottom(model)
									])),
								A2(
								elm$html$Html$main_,
								_List_fromArray(
									[
										elm$html$Html$Attributes$class('col')
									]),
								model.serverIsOnline ? _List_fromArray(
									[
										A2(author$project$Main$viewCurrentRoute, model, viewState),
										A2(
										elm$html$Html$map,
										author$project$Main$ListMsg,
										author$project$Routes$Ls$buildModals(viewState.listState)),
										A2(
										elm$html$Html$map,
										author$project$Main$RemotesMsg,
										author$project$Routes$Remotes$buildModals(viewState.remoteState))
									]) : _List_fromArray(
									[author$project$Main$viewOfflineMarker]))
							]))
					]))
			]);
	});
var author$project$Main$view = function (model) {
	return {
		body: function () {
			var _n0 = model.loginState;
			switch (_n0.$) {
				case 'LoginLimbo':
					return _List_fromArray(
						[
							elm$html$Html$text('Waiting for login data')
						]);
				case 'LoginReady':
					return _List_fromArray(
						[
							A2(elm$html$Html$Lazy$lazy, author$project$Main$viewLoginForm, model)
						]);
				case 'LoginFailure':
					return _List_fromArray(
						[
							A2(elm$html$Html$Lazy$lazy, author$project$Main$viewLoginForm, model)
						]);
				case 'LoginLoading':
					return _List_fromArray(
						[
							A2(elm$html$Html$Lazy$lazy, author$project$Main$viewLoginForm, model)
						]);
				default:
					var viewState = _n0.a;
					return A2(author$project$Main$viewMainContent, model, viewState);
			}
		}(),
		title: 'Gateway'
	};
};
var elm$browser$Browser$application = _Browser_application;
var author$project$Main$main = elm$browser$Browser$application(
	{init: author$project$Main$init, onUrlChange: author$project$Main$UrlChanged, onUrlRequest: author$project$Main$LinkClicked, subscriptions: author$project$Main$subscriptions, update: author$project$Main$update, view: author$project$Main$view});
_Platform_export({'Main':{'init':author$project$Main$main(
	elm$json$Json$Decode$succeed(_Utils_Tuple0))(0)}});}(this));