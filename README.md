# simia

<img src="./img/simia.png" width="100" />

> Created with https://www.craiyon.com/ and manually vectorized in Figma

[![CI](https://github.com/protiumx/simia/actions/workflows/ci.yml/badge.svg)](https://github.com/protiumx/simia/actions/workflows/ci.yml)

Go implementation of the Monkey language interpreter from the book [Writing an interpreter in Go](https://interpreterbook.com/)

> *NOTE*: this repo is work in progress

## Extending Monkey
I have extended and took some different decisions while designing the `simia` language

- No `nil` value supported. Variables must always be defined with the supported types
- Added `Range` support defined by start and end integers
- Added support for `in` operator
- Added `for-loop` support with boolean or `in` expressions
- Parentheses are optional for `if` and `for` blocks
- Added `|>` operator (borrowed from [elixir](https://elixirschool.com/en/lessons/basics/pipe_operator))

## Usage
Run the REPL
```go
go run main
```

## Syntax
### Variables declaration and assignment
```
let foo = "";
foo = "bar";
```

### Arithmetic expressions
```
let a = 13;
let b = 19 * a;
b = (b / 20) * a;
```

### For loops
```
for i in 1..10 {
  log(i);
}

let i = 0;
for i < 10 {
  log(i);
}

let ret = ""
for el in ["hello", "universe"] {
  ret = ret + el
}
```

### Builtin functions
- `len(<iterable>)`: Returns length of iterable (string, array, range)
- `log(...args)`: Prints arguments to the standard output followed by a new line
- `append(array)`: Pushes value to the end of the array

### Types
Type      | Syntax                                    
--------- | -----------------------------------------
bool      | `true` | `false`                         
int       | `0 42 1234 -5`                           
string    | `"" "foo" "\"quotes\" and a\nline break"`
array     | `[] [1, 2] [1, 2, 3]`                    
hash      | `{} {"a": 1} {"a": 1, "b": 2}`         

## TODO
- [ ] Add `collumn` and `line` numbers
- [ ] Implement `array` with spread `...` operator
- [x] Support piping like in Elixir (`|>`)

## License

MIT License

Copyright (c) 2022 Brian Mayo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
