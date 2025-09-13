# THE DOLME PROGRAMMING LANGUAGE

<p align="center">
<img src="https://github.com/uidops/uidops/blob/main/Dolmas+colour.jpg?raw=true">
</p>

DOLME is a student project for a programming language for compiler desiging course at Azarbaijan Shahid Madani University written in Go.
The main goal of this project is to learn how to design and implement a programming language and its compiler.
I used LL(1) parsing technique and Syntax-Directed Translation to implement the compiler.

Currently arm64-macos are supported for assembly generation.

Professor provided us with a simple grammar and we extended it to support more features like functions and some operations.

The main grammer and problem of the professor can be found in `dolme.pdf`

Context-Free Grammar of DOLME and LL(1) Parsing table can be found in `dolme.md`



## Example:
```js
func pow(a: float, b: int): float {
    let c : float = 1.0;
    while (b > 0) {
        c = c * a;
        b = b - 1;
    }
    return c;
}

func factorial(a: int): float {
    let c : float = 1.0;
    while (a > 0) {
        c = c * a;
        a = a - 1;
    }
    return c;
}


func sin(x: float): float {
    let y : float = x;
    let e : int = 3;
    let i : int = 1;

    while (i < 50) {
        y = y + (pow(-1.0, i) * (pow(x, e) / factorial(e)));
        e = e + 2;
        i = i + 1;
    }

    return y;
}

let b : float = sin(4.4249);
print(b);
```

## Usage:
```bash
$ make # or go build -o bin/dolme cmd/main.go
$ bin/dolme -h
$ bin/dolme -v examples/01.dolme # too see generated IR code
$ bin/dolme -r examples/01.dolme # run in interpreter mode
$ bin/dolme -c -a arm64-macos examples/01.dolme # compile to binary
$ bin/dolme -c -a arm64-macos -v examples/01.dolme # compiler to binary and show generated assembly
```
