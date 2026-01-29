# F1-Lang Language Specification

F1-Lang is a Formula 1-themed programming language with a modular compiler that produces native executables via LLVM.

---

## Table of Contents

1. [Lexical Structure](#1-lexical-structure)
2. [Type System](#2-type-system)
3. [Grammar](#3-grammar)
4. [Semantics](#4-semantics)
5. [Built-in Functions](#5-built-in-functions)
6. [Scoping Rules](#6-scoping-rules)
7. [Design Decisions](#7-design-decisions)
8. [Compiler Architecture](#8-compiler-architecture)
9. [F1-IR Specification](#9-f1-ir-specification)
10. [Optimizer Passes](#10-optimizer-passes)
11. [LLVM Code Generation](#11-llvm-code-generation)
12. [Example Programs](#12-example-programs)

---

## 1. Lexical Structure

### 1.1 Comments

```
// Single-line comments start with double slash
```

### 1.2 Keywords

| Keyword | Standard Equivalent | Description |
|---------|---------------------|-------------|
| `driver` | `var` | Variable declaration |
| `pitstop` | `func` | Function declaration |
| `drs` | `if` | Conditional statement |
| `defend` | `else` | Else clause |
| `lap` | `for` | Loop statement |
| `finish` | `return` | Return statement |
| `greenlight` | `true` | Boolean true |
| `redlight` | `false` | Boolean false |

### 1.3 Operators

**Arithmetic:**
| Operator | Description | Precedence |
|----------|-------------|------------|
| `+` | Addition | 5 |
| `-` | Subtraction | 5 |
| `*` | Multiplication | 6 |
| `/` | Division | 6 |
| `%` | Modulo | 6 |

**Comparison:**
| Operator | Description | Precedence |
|----------|-------------|------------|
| `==` | Equal | 3 |
| `!=` | Not equal | 3 |
| `<` | Less than | 4 |
| `>` | Greater than | 4 |
| `<=` | Less or equal | 4 |
| `>=` | Greater or equal | 4 |

**Logical:**
| Operator | Description | Precedence |
|----------|-------------|------------|
| `!` | Logical NOT | 7 (unary) |
| `&&` | Logical AND | 2 |
| `\|\|` | Logical OR | 1 |

**Assignment:**
| Operator | Description |
|----------|-------------|
| `=` | Assignment |

### 1.4 Delimiters

| Symbol | Usage |
|--------|-------|
| `(` `)` | Grouping, function calls, parameters |
| `{` `}` | Blocks |
| `;` | Statement terminator |
| `,` | Separator |

### 1.5 Literals

**Integer literals:**
```
0
42
1000000
```

**Integer literal constraints:**
- Must fit in signed 64-bit range: `-9223372036854775808` to `9223372036854775807`
- Overflow in literal → **Lexer error**

Note: Negative numbers like `-17` are parsed as unary minus applied to `17`, not as a single token.

**String literals:**
```
"Hello, World!"
"Lights out and away we go!"
""
```

**Escape sequences (within strings):**
| Sequence | Meaning |
|----------|---------|
| `\\` | Backslash |
| `\"` | Double quote |
| `\n` | Newline |
| `\t` | Tab |

Example: `"Line 1\nLine 2"`, `"He said \"hello\""`

**Invalid escape sequences:** Any `\` followed by a character not in the table above is a **Lexer error**.
```
"hello\q"    // ERROR: invalid escape sequence '\q'
```

**Boolean literals:**
```
greenlight   // true
redlight     // false
```

### 1.6 Identifiers

```
[a-zA-Z_][a-zA-Z0-9_]*
```

Examples: `x`, `points`, `max_speed`, `driver1`, `_private`

### 1.7 Token Types

```
// Literals
TOKEN_INT           // 42
TOKEN_STRING        // "hello"
TOKEN_IDENT         // myVar

// Keywords
TOKEN_DRIVER        // driver
TOKEN_PITSTOP       // pitstop
TOKEN_DRS           // drs
TOKEN_DEFEND        // defend
TOKEN_LAP           // lap
TOKEN_FINISH        // finish
TOKEN_GREENLIGHT    // greenlight
TOKEN_REDLIGHT      // redlight

// Operators
TOKEN_PLUS          // +
TOKEN_MINUS         // -
TOKEN_STAR          // *
TOKEN_SLASH         // /
TOKEN_PERCENT       // %
TOKEN_EQ            // ==
TOKEN_NEQ           // !=
TOKEN_LT            // <
TOKEN_GT            // >
TOKEN_LTE           // <=
TOKEN_GTE           // >=
TOKEN_AND           // &&
TOKEN_OR            // ||
TOKEN_NOT           // !
TOKEN_ASSIGN        // =

// Delimiters
TOKEN_LPAREN        // (
TOKEN_RPAREN        // )
TOKEN_LBRACE        // {
TOKEN_RBRACE        // }
TOKEN_SEMICOLON     // ;
TOKEN_COMMA         // ,

// Special
TOKEN_EOF           // end of file
TOKEN_ILLEGAL       // unrecognized character
```

---

## 2. Type System

### 2.1 Primitive Types

| Type | Size | Description | Default |
|------|------|-------------|---------|
| `int` | 64-bit | Signed integer | `0` |
| `bool` | 1-bit | Boolean | `redlight` |
| `string` | ptr | UTF-8 string | `""` |
| `void` | 0 | No value (function returns) | N/A |
| `function` | ptr | First-class function (closure) | N/A |

### 2.2 Type Inference Rules

Variables infer their type from initialization:

```
driver x = 42;           // x : int
driver s = "hello";      // s : string
driver b = greenlight;   // b : bool
driver r = add(1, 2);    // r : <return type of add>
```

### 2.3 Type Compatibility

| Operation | Left Type | Right Type | Result Type |
|-----------|-----------|------------|-------------|
| `+` `-` `*` `/` `%` | `int` | `int` | `int` |
| `<` `>` `<=` `>=` | `int` | `int` | `bool` |
| `==` `!=` | `int` | `int` | `bool` |
| `==` `!=` | `bool` | `bool` | `bool` |
| `==` `!=` | `string` | `string` | `bool` |
| `&&` `\|\|` | `bool` | `bool` | `bool` |
| `!` | `bool` | N/A | `bool` |
| `-` (unary) | `int` | N/A | `int` |

**Notes:**
- String concatenation (`string + string`) is NOT supported in v1. Use multiple `radio()` calls instead.
- String ordering (`<`, `>`, `<=`, `>=`) is NOT supported. Only `==` and `!=` work for strings.

### 2.4 Type Errors

The type checker reports errors for:

1. **Type mismatch:** `driver x = 5 + "hello";`
2. **Undefined variable:** `radio(undefined_var);`
3. **Undefined function:** `unknown_func();`
4. **Wrong argument count:** `add(1);` when `add` takes 2 params
5. **Wrong argument type:** `add("a", "b");` when `add` takes ints
6. **Non-boolean condition:** `drs (42) { ... }`
7. **Return type mismatch:** returning `string` from `int` function
8. **Redeclaration in same scope:** `driver x = 1; driver x = 2;`
9. **Reserved identifier:** `driver radio = 5;` or `pitstop main() { }`
10. **Assignment type mismatch:** `driver x = 5; x = "hello";`
11. **Void in expression:** `driver x = voidFunc();` — void values cannot be assigned or used in expressions
12. **Wrong argument count (graphics):** `pixel(0, 0);` — pixel requires 5 arguments
13. **Non-callable type:** `driver x = 5; x();` — cannot call non-function value
14. **Closure type mismatch:** `driver f = pitstop(driver x) { finish x; }; f("hi");` — wrong argument type for closure
15. **Function comparison:** `f == g` where f or g is a closure — cannot compare function types

**Note:** Using `pixel()` before `canvas()` is a **runtime error**, not a type error (see §7.17).

---

## 3. Grammar

### 3.1 Program Structure

```ebnf
Program = { TopLevel } EOF ;

TopLevel = FunctionDecl | Statement ;
```

### 3.2 Statements

```ebnf
Statement = VariableDecl
          | Assignment
          | IfStatement
          | LoopStatement
          | ReturnStatement
          | ExpressionStatement
          | Block ;

VariableDecl = "driver" IDENTIFIER "=" Expression ";" ;

Assignment = IDENTIFIER "=" Expression ";" ;

FunctionDecl = "pitstop" IDENTIFIER "(" [ ParameterList ] ")" Block ;

ParameterList = Parameter { "," Parameter } ;
Parameter = "driver" IDENTIFIER ;

IfStatement = "drs" "(" Expression ")" Block [ "defend" ( IfStatement | Block ) ] ;

LoopStatement = "lap" "(" LoopInit LoopCond ";" LoopUpdate ")" Block ;
LoopInit = "driver" IDENTIFIER "=" Expression ";" ;
LoopCond = Expression ;
LoopUpdate = IDENTIFIER "=" Expression ;   /* Note: no semicolon */

ReturnStatement = "finish" [ Expression ] ";" ;

ExpressionStatement = Expression ";" ;

Block = "{" { Statement } "}" ;
```

**Loop syntax clarification:**
```
lap (driver i = 0; i < 10; i = i + 1) { ... }
     ^^^^^^^^^^^^  ^^^^^^  ^^^^^^^^^^
     LoopInit      Cond    Update (no trailing semicolon)
```

### 3.3 Expressions

```ebnf
Expression = LogicalOr ;

LogicalOr = LogicalAnd { "||" LogicalAnd } ;

LogicalAnd = Equality { "&&" Equality } ;

Equality = Comparison { ( "==" | "!=" ) Comparison } ;

Comparison = Term { ( "<" | ">" | "<=" | ">=" ) Term } ;

Term = Factor { ( "+" | "-" ) Factor } ;

Factor = Unary { ( "*" | "/" | "%" ) Unary } ;

Unary = ( "!" | "-" ) Unary | Call ;

Call = Primary { "(" [ ArgumentList ] ")" } ;

ArgumentList = Expression { "," Expression } ;

Primary = INTEGER
        | STRING
        | "greenlight"
        | "redlight"
        | IDENTIFIER
        | "(" Expression ")"
        | LambdaExpr ;

LambdaExpr = "pitstop" "(" [ ParameterList ] ")" Block ;
```

### 3.4 Operator Precedence Table

| Level | Operators | Associativity | Description |
|-------|-----------|---------------|-------------|
| 1 | `\|\|` | Left | Logical OR |
| 2 | `&&` | Left | Logical AND |
| 3 | `==` `!=` | Left | Equality |
| 4 | `<` `>` `<=` `>=` | Left | Comparison |
| 5 | `+` `-` | Left | Additive |
| 6 | `*` `/` `%` | Left | Multiplicative |
| 7 | `!` `-` | Right | Unary |
| 8 | `()` | Left | Call |

---

## 4. Semantics

### 4.1 Variable Declaration

Variables must be initialized:

```
driver points = 25;
driver name = "Verstappen";
driver active = greenlight;
```

### 4.2 Assignment

Variables can be reassigned (same type only):

```
driver x = 10;
x = x + 1;      // OK
x = "hello";    // ERROR: type mismatch
```

### 4.3 Functions

```
pitstop add(driver a, driver b) {
    finish a + b;
}

driver result = add(3, 4);  // result = 7
```

- Parameters are pass-by-value
- Recursion is supported
- Functions must return a consistent type

### 4.4 Control Flow

**If-Else:**
```
drs (condition) {
    // then branch
} defend {
    // else branch (optional)
}
```

**Chained conditions:**
```
drs (x < 0) {
    radio("negative");
} defend drs (x == 0) {
    radio("zero");
} defend {
    radio("positive");
}
```

**Loops:**
```
lap (driver i = 0; i < 10; i = i + 1) {
    radio(i);
}
```

### 4.5 Return Statements

```
pitstop early(driver n) {
    drs (n < 0) {
        finish 0;  // early return
    }
    finish n * 2;
}
```

### 4.6 First-Class Functions (Closures)

Functions can be assigned to variables and passed as arguments:

```
// Lambda assigned to variable
driver addOne = pitstop(driver x) { finish x + 1; };
radio(addOne(5));  // 6

// Closure capturing outer variable
driver multiplier = 10;
driver scale = pitstop(driver x) { finish x * multiplier; };
radio(scale(5));  // 50

// Higher-order function
pitstop apply(driver f, driver x) {
    finish f(x);
}
radio(apply(addOne, 10));  // 11
```

**Capture semantics:**
- Variables are captured **by value** at lambda creation time
- Mutations to captured variables inside the closure do NOT affect the outer scope
- Mutations to outer variables after closure creation do NOT affect captured values

```
driver x = 10;
driver f = pitstop() { finish x; };  // captures x = 10
x = 20;
radio(f());  // 10 (captured value, not current)
```

### 4.7 Tail Call Optimization

Tail-recursive functions are optimized to use constant stack space:

```
// Tail-recursive fibonacci (optimized)
pitstop fib(driver n, driver a, driver b) {
    drs (n == 0) { finish a; }
    finish fib(n - 1, b, a + b);  // tail position → becomes jump
}

radio(fib(50, 0, 1));  // no stack overflow
```

**Tail position rules:**
- A call is in tail position if it's the last action before `finish`
- `finish f(x);` — tail call
- `finish f(x) + 1;` — NOT tail call (addition happens after)
- `finish drs (c) { f(x); } defend { g(y); };` — NOT supported (no ternary)

---

## 5. Built-in Functions

### 5.1 `radio(value)`

Prints to stdout with newline.

| Argument Type | Output |
|---------------|--------|
| `int` | Decimal number |
| `string` | String content |
| `bool` | `true` or `false` |

```
radio("Lights out!");    // Lights out!
radio(42);               // 42
radio(greenlight);       // true
```

### 5.2 `bono(message)`

Prints error to stderr with F1 meme prefix.

```
bono("critical wear");
// Output: Bono, my tyres are gone! critical wear
```

### 5.3 Graphics Functions

F1-Lang includes built-in graphics functions for generating PPM images.

#### `canvas(width, height)`

Initializes an internal pixel buffer.

| Parameter | Type | Description |
|-----------|------|-------------|
| `width` | `int` | Canvas width in pixels (clamped to 1-4096) |
| `height` | `int` | Canvas height in pixels (clamped to 1-4096) |
| Returns | `void` | |

- Calling `canvas` clears any existing buffer
- All pixels initialize to black (0, 0, 0)
- Must be called before `pixel` or `render`

#### `pixel(x, y, r, g, b)`

Sets a pixel color in the canvas buffer.

| Parameter | Type | Description |
|-----------|------|-------------|
| `x` | `int` | X coordinate (0 = left) |
| `y` | `int` | Y coordinate (0 = top) |
| `r` | `int` | Red component (0-255) |
| `g` | `int` | Green component (0-255) |
| `b` | `int` | Blue component (0-255) |
| Returns | `void` | |

- Coordinates outside canvas bounds are ignored (no error)
- Color values are clamped to 0-255

#### `render(filename)`

Writes the canvas buffer to a PPM file.

| Parameter | Type | Description |
|-----------|------|-------------|
| `filename` | `string` | Output file path (e.g., `"output.ppm"`) |
| Returns | `void` | |

- Writes P6 binary PPM format
- Overwrites existing file if present
- Conventionally use `.ppm` extension (not enforced)

#### `snapshot(frame_number)`

Writes the canvas buffer to a numbered PPM file for animations.

| Parameter | Type | Description |
|-----------|------|-------------|
| `frame_number` | `int` | Frame number (any int; 0-9999 recommended for 4-digit padding) |
| Returns | `void` | |

- Writes to `frame_NNNN.ppm` (e.g., `frame_0042.ppm`)
- Negative values produce `frame_-NNN.ppm`; values >9999 use more digits
- Use with ffmpeg: `ffmpeg -i frame_%04d.ppm output.gif`

---

## 6. Scoping Rules

### 6.1 Block Scope

Variables are scoped to their enclosing block:

```
driver x = 10;
drs (greenlight) {
    driver x = 20;   // shadows outer x
    radio(x);        // 20
}
radio(x);            // 10
```

**Standalone blocks** are allowed and create a new scope:

```
driver x = 10;
{
    driver x = 20;   // shadows outer x
    radio(x);        // 20
}
radio(x);            // 10
```

This is useful for limiting variable lifetime.

### 6.2 Function Scope

- Functions are globally visible (hoisted)
- Parameters are local to the function body
- Variables declared in function body are local

### 6.3 Symbol Table Structure

```
GlobalScope
├── functions: { "add": FunctionType, "fib": FunctionType, ... }
├── builtins: { "radio", "bono", "canvas", "pixel", "render", "snapshot" }
└── children:
    ├── FunctionScope("add")
    │   ├── variables: { "a": int, "b": int }
    │   └── children: [BlockScope, ...]
    └── FunctionScope("fib")
        └── ...
```

---

## 7. Design Decisions

This section documents explicit design decisions for edge cases and ambiguities.

### 7.1 Entry Point

**Decision:** Top-level statements are wrapped in an implicit `@main` function.

```
// Source:
radio("Hello");

// Becomes:
define i32 @main() {
    call @radio("Hello")
    ret i32 0
}
```

**Reserved identifier:** `main` is reserved. User cannot define `pitstop main()`.

**Top-level variables:** Variables declared at top level become `alloca` instructions inside `@main`, not global variables.

```
// Source:
driver x = 42;
radio(x);

// IR (inside @main):
%x = alloca i64
store i64 42, %x
%0 = load i64, %x
call void @radio_int(%0)
```

### 7.2 Function Return Types

**Decision:** Return types are inferred from `finish` statements.

**Rules:**
1. If function has no `finish` statement → return type is `void`
2. If function has `finish;` (no value) → return type is `void`
3. If function has `finish <expr>;` → return type is type of `<expr>`
4. All `finish` statements in a function must return the same type (**error** otherwise)
5. Functions with non-void return type should have a `finish` on all code paths (**warning**, not error)
   - If control reaches end of non-void function, behavior is undefined (returns garbage)
   - Full control flow analysis is complex; we emit a warning but don't enforce

**Rationale for Rule 5:** Precise "return on all paths" analysis requires solving the halting problem in general. We warn but allow compilation. Matches C behavior.

**Examples:**
```
pitstop greet() {
    radio("Hi");
}   // void return (implicit finish)

pitstop add(driver a, driver b) {
    finish a + b;
}   // int return

pitstop bad(driver x) {
    drs (x > 0) {
        finish x;
    }
    finish "error";   // ERROR: inconsistent return types
}
```

### 7.3 Implicit Return for Void Functions

**Decision:** Void functions without an explicit `finish` get an implicit `finish;` at the end.

```
pitstop noop() { }
// Equivalent to:
pitstop noop() { finish; }

pitstop greet() {
    radio("Hello");
}
// Equivalent to:
pitstop greet() {
    radio("Hello");
    finish;
}
```

**Implementation:** During IR generation, if a void function's last block doesn't end with a `ret`, insert `ret void`.

### 7.4 Short-Circuit Evaluation

**Decision:** `&&` and `||` use short-circuit evaluation (like C/Go/JavaScript).

```
// Right side NOT evaluated if left side determines result
redlight && expensiveCall()    // expensiveCall() not called
greenlight || expensiveCall()  // expensiveCall() not called
```

**IR implementation:** Conditional branches, not eager evaluation.

### 7.5 Negative Numbers

**Decision:** Negative numbers are parsed as unary minus, not lexer literals.

```
-42  →  UnaryExpr { Op: "-", Arg: IntLit(42) }
```

**Rationale:** Simpler lexer. Parser handles precedence correctly.

### 7.6 Integer Overflow

**Decision:** Wrap around (64-bit two's complement). No runtime error.

```
driver x = 9223372036854775807;  // max int64
x = x + 1;                        // wraps to -9223372036854775808
```

**Rationale:** Match LLVM/C behavior. Predictable, fast.

### 7.7 Division by Zero

**Decision:** Undefined behavior (follows LLVM semantics).

```
driver x = 10 / 0;   // undefined - may crash, may return garbage
```

**Rationale:** Checking adds runtime overhead. Matches C behavior. User's responsibility.

### 7.8 Built-in Function Typing

**Decision:** Built-ins are compiler magic, not user-definable function types.

The type checker handles built-ins specially:

```
// I/O
radio(int)     → void    // prints integer
radio(string)  → void    // prints string
radio(bool)    → void    // prints "true" or "false"
bono(string)   → void    // prints error message to stderr

// Graphics
canvas(int, int)              → void    // initialize canvas
pixel(int, int, int, int, int) → void   // set pixel (x, y, r, g, b)
render(string)                → void    // write to PPM file
snapshot(int)                 → void    // write numbered frame
```

**Implementation:** Type checker recognizes built-in names and validates arguments directly, rather than looking them up in symbol table as regular functions.

### 7.9 Shadowing Rules

**Decision:** Variables can shadow outer scope variables AND function parameters, but NOT in the same scope.

```
pitstop example(driver x) {
    driver x = 10;     // shadows parameter x (allowed - different scope)
    radio(x);          // prints 10
}

// Same-scope redeclaration is an ERROR:
driver x = 5;
driver x = 10;         // ERROR: 'x' already declared in this scope
```

**Rationale:** Consistent with block scoping. Matches Go/Rust behavior.

### 7.10 Phi Node Generation

**Decision:** Phi nodes are generated ONLY for loops (for the loop variable).

If-else does NOT require phi nodes because:
- Each branch terminates with its own instructions
- No value needs to "merge" after if-else (we don't have ternary expressions)

**Loop phi example:**
```
lap (driver i = 0; i < 10; i = i + 1) { ... }

// IR:
loop.header:
    %i = phi i64 [0, %entry], [%i.next, %loop.body]
    ...
loop.body:
    %i.next = add i64 %i, 1
    br %loop.header
```

### 7.11 String Equality

**Decision:** String `==` and `!=` compare by value (content), not reference.

```
driver a = "hello";
driver b = "hello";
radio(a == b);       // true (same content)
```

**LLVM implementation:** Call `strcmp` and check result.

### 7.12 Reserved Identifiers

The following identifiers are reserved and cannot be used as variable/function names:

| Reserved | Reason |
|----------|--------|
| `main` | Entry point |
| `radio` | Built-in function |
| `bono` | Built-in function |
| `canvas` | Built-in function |
| `pixel` | Built-in function |
| `render` | Built-in function |
| `snapshot` | Built-in function |
| `driver` | Keyword |
| `pitstop` | Keyword |
| `drs` | Keyword |
| `defend` | Keyword |
| `lap` | Keyword |
| `finish` | Keyword |
| `greenlight` | Keyword |
| `redlight` | Keyword |

**Note:** Built-ins (`radio`, `bono`) cannot be shadowed by variables. This is a **compile-time error**:
```
driver radio = 5;      // ERROR: 'radio' is a reserved identifier
pitstop bono() { }     // ERROR: 'bono' is a reserved identifier
```

### 7.13 Comments

**Decision:** Single-line comments only (`//`). No multi-line comments.

```
// This is a comment
driver x = 5;  // inline comment
```

**Rationale:** Simpler lexer. Multi-line comments can be added in v2.

### 7.14 Maximum Identifier Length

**Decision:** No limit (bounded by available memory).

### 7.15 Unicode Support

**Decision:**
- Source files: ASCII only for keywords/identifiers
- Strings: UTF-8 content allowed

```
driver emoji = "🏎️";    // OK in string
driver 变量 = 5;         // ERROR: non-ASCII identifier
```

### 7.16 Void Type Restrictions

**Decision:** `void` is not a first-class type. It cannot be used in expressions or assigned to variables.

**Illegal operations:**
```
pitstop nothing() { }

driver x = nothing();           // ERROR: cannot assign void
driver y = 5 + nothing();       // ERROR: cannot use void in expression
radio(nothing());               // ERROR: cannot pass void as argument
```

**Legal operations:**
```
nothing();                      // OK: void function call as statement
pitstop wrapper() {
    nothing();                  // OK: void call inside void function
}
```

**Rationale:** `void` represents absence of value. Allowing void in expressions would require special handling everywhere and adds complexity with no benefit.

### 7.17 Graphics Buffer

**Decision:** The canvas is a single global buffer managed internally by the runtime.

**Rules:**
1. Only one canvas exists at a time
2. `canvas()` resets and reallocates the buffer
3. Buffer is stored as contiguous RGB bytes (row-major order)
4. Maximum canvas size: 4096 × 4096 pixels (values clamped: <1 becomes 1, >4096 becomes 4096)
5. `pixel()` before `canvas()` is a **runtime error** (program terminates)

**Rationale:** Keeping graphics simple avoids need for arrays, pointers, or memory management in the language.

### 7.18 Graphics Coordinate System

**Decision:** Origin (0, 0) is top-left corner.

```
(0,0) ────────→ X
  │
  │
  ↓
  Y
```

**Rationale:** Matches PPM format and common image coordinate conventions.

### 7.19 Graphics Value Clamping

**Decision:** Color values and coordinates are clamped silently (no error).

| Value | Behavior |
|-------|----------|
| `r, g, b < 0` | Clamped to 0 |
| `r, g, b > 255` | Clamped to 255 |
| `x < 0` or `x >= width` | Pixel ignored |
| `y < 0` or `y >= height` | Pixel ignored |

**Rationale:** Simplifies drawing code — no bounds checking needed by user.

### 7.20 Closure Representation

**Decision:** Closures are represented as a pair: (function pointer, environment pointer).

**Structure:**
```
Closure = { ptr function, ptr environment }
Environment = struct { captured_var_1, captured_var_2, ... }
```

**Example:**
```
driver x = 10;
driver y = 20;
driver f = pitstop(driver z) { finish x + y + z; };
```

**Generated representation:**
```
// Environment struct (heap allocated)
%env = struct { i64 10, i64 20 }   // captures x, y

// Closure struct
%f = struct { ptr @lambda.0, ptr %env }

// Lambda function (takes hidden env parameter)
define i64 @lambda.0(ptr %env, i64 %z) {
    %x = load i64, getelementptr %env, 0
    %y = load i64, getelementptr %env, 1
    %sum = add i64 %x, %y
    %result = add i64 %sum, %z
    ret i64 %result
}
```

**Calling convention:**
- When calling a closure, extract function pointer and environment
- Pass environment as hidden first argument

**Memory management:**
- Environments are heap-allocated (malloc)
- No garbage collection in v1 — environments leak
- Future: add reference counting or tracing GC

### 7.21 Free Variable Analysis

**Decision:** The compiler performs free variable analysis to determine what each lambda captures.

**Algorithm:**
1. Walk the lambda body, collecting all variable references
2. For each reference, check if it's a parameter or locally declared
3. If neither, it's a free variable that must be captured

**Example:**
```
driver a = 1;
driver b = 2;
driver f = pitstop(driver x) {
    driver y = 3;
    finish a + x + y;  // 'a' is free, 'x' is param, 'y' is local
};
// f captures: { a }
```

**Nested lambdas:** Free variables propagate through nesting:
```
driver x = 10;
driver f = pitstop() {
    driver g = pitstop() {
        finish x;  // x is free in g AND in f
    };
    finish g();
};
// f captures: { x }
// g captures: { x } (from f's environment)
```

### 7.22 Tail Call Optimization

**Decision:** Tail calls are detected and optimized to eliminate stack growth.

**Detection rules:**
A call `f(...)` is in tail position if:
1. It appears as `finish f(...);` with no operations after the call
2. The return type of `f` matches the enclosing function's return type

**NOT tail calls:**
```
finish f(x) + 1;       // addition after call
finish f(x) * g(y);    // multiplication after call
driver r = f(x);       // assigned to variable
finish r;              // not direct
```

**LLVM implementation:**
- Mark tail calls with `tail` or `musttail` attribute
- LLVM backend converts to jumps instead of calls

```llvm
; Tail-recursive call
tail call i64 @fib(i64 %n.next, i64 %b, i64 %sum)
ret i64 %result   ; never reached, optimized away
```

**Self-recursion only:** In v1, we only optimize direct self-recursive tail calls. Mutual recursion (A calls B, B calls A in tail position) is not optimized.

### 7.23 Closure Type Checking

**Decision:** Closure types are structural, not nominal.

Two closure types are compatible if:
1. They have the same parameter types (in order)
2. They have the same return type

```
driver f = pitstop(driver x) { finish x + 1; };
driver g = pitstop(driver y) { finish y * 2; };

// f and g have the same type: (int) -> int
f = g;  // OK - same type
```

**Passing closures:**
```
pitstop apply(driver fn, driver x) {
    finish fn(x);
}

driver double = pitstop(driver n) { finish n * 2; };
radio(apply(double, 5));  // 10
```

**Type inference for parameters:**
- Parameters accepting closures infer type from usage
- If `fn(x)` is called with int `x` and result is used as int, then `fn: (int) -> int`

### 7.24 Closure Equality

**Decision:** Closure comparison (`==`, `!=`) is NOT supported.

```
driver f = pitstop(driver x) { finish x; };
driver g = pitstop(driver x) { finish x; };
radio(f == g);    // ERROR: cannot compare function types
```

**Rationale:** Two closures with identical code but different environments are semantically different. Comparing function pointers alone is misleading. This is a compile-time type error.

---

## 8. Compiler Architecture

### 8.1 Pipeline Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                      │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│   Source (.f1)                                                        │
│       │                                                               │
│       ▼                                                               │
│   ┌────────┐     ┌────────┐     ┌─────────────┐                      │
│   │ Lexer  │────▶│ Parser │────▶│ TypeChecker │                      │
│   └────────┘     └────────┘     └─────────────┘                      │
│       │              │                │                               │
│       ▼              ▼                ▼                               │
│   []Token          AST           TypedAST                            │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         MIDDLE-END                                    │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│   ┌─────────┐     ┌───────────┐                                      │
│   │  IRGen  │────▶│ Optimizer │                                      │
│   └─────────┘     └───────────┘                                      │
│       │                │                                              │
│       ▼                ▼                                              │
│    F1-IR          Optimized F1-IR                                    │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         BACKEND                                       │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│   ┌─────────┐     ┌───────┐     ┌────────┐                           │
│   │ Codegen │────▶│ clang │────▶│ Binary │                           │
│   └─────────┘     └───────┘     └────────┘                           │
│       │                                                               │
│       ▼                                                               │
│   LLVM IR (.ll)                                                       │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### 8.2 Module Contracts

| Module | Input | Output | Errors |
|--------|-------|--------|--------|
| Lexer | Source code | Token stream | Lexical errors |
| Parser | Token stream | AST | Syntax errors |
| TypeChecker | AST | Typed AST | Type errors |
| IRGen | Typed AST | F1-IR | None |
| Optimizer | F1-IR | Optimized F1-IR | None |
| Codegen | Optimized F1-IR | LLVM IR | None |
| Runtime | — | Object file | — |

**Note:** The runtime library provides graphics functions and is linked with compiled programs.

---

## 9. F1-IR Specification

F1-IR is our **custom intermediate representation**, distinct from LLVM IR. It has a simpler syntax optimized for our compiler.

### 9.1 Design Principles

1. **SSA Form:** Each virtual register assigned exactly once
2. **Three-Address Code:** At most 3 operands per instruction
3. **Explicit Control Flow:** Basic blocks with explicit terminators
4. **Typed Values:** Every value carries its type
5. **Simple Syntax:** Labels are bare names (not `%label` or `label:`), branches use `br cond, then, else`

### 9.2 Module Structure

```
Module
├── Globals[]          // Global string constants
└── Functions[]
    └── Function
        ├── Name
        ├── Params[]
        ├── ReturnType
        └── Blocks[]
            └── BasicBlock
                ├── Label
                ├── Instructions[]
                └── Terminator
```

### 9.3 Values

```
Value = Constant | VirtualReg | Parameter

Constant:
  - IntConst(value: int64)
  - BoolConst(value: bool)
  - StringConst(index: int)    // reference to global

VirtualReg:
  - %0, %1, %2, ...            // SSA temporaries

Parameter:
  - %param0, %param1, ...      // function parameters
```

### 9.4 Instructions

**Memory:**
```
%r = alloca <type>              // allocate stack space
%r = load <type>, %ptr          // load from memory
store <type> %val, %ptr         // store to memory
```

**Arithmetic:**
```
%r = add <type> %a, %b          // addition
%r = sub <type> %a, %b          // subtraction
%r = mul <type> %a, %b          // multiplication
%r = div <type> %a, %b          // division
%r = mod <type> %a, %b          // modulo
%r = neg <type> %a              // unary negation
```

**Comparison:**
```
%r = eq <type> %a, %b           // equal
%r = neq <type> %a, %b          // not equal
%r = lt <type> %a, %b           // less than
%r = gt <type> %a, %b           // greater than
%r = lte <type> %a, %b          // less or equal
%r = gte <type> %a, %b          // greater or equal
```

**Logical:**
```
%r = and bool %a, %b            // logical AND
%r = or bool %a, %b             // logical OR
%r = not bool %a                // logical NOT
```

**Function:**
```
%r = call <ret_type> @func(%arg1, %arg2, ...)
call void @func(%arg1, ...)     // void call (no result)
```

**Miscellaneous:**
```
%r = copy <type> %val           // copy value (for SSA)
%r = phi <type> [%v1, %bb1], [%v2, %bb2]  // SSA phi node
```

**Closure:**
```
%env = alloc_env [<type>, ...]   // allocate environment struct
store_env %env, <index>, %val    // store captured value at index
%r = load_env %env, <index>      // load captured value from index
%r = make_closure @func, %env    // create closure (func ptr + env ptr)
%r = closure_call %closure, [%arg1, ...]  // call a closure
```

### 9.5 Terminators

Every basic block ends with exactly one terminator:

```
ret <type> %val                 // return with value
ret void                        // return nothing

br %label                       // unconditional branch

br %cond, %then_label, %else_label  // conditional branch

tailcall <ret_type> @func(%args...)  // tail call (replaces ret)
```

### 9.6 Example IR

**Source:**
```
pitstop max(driver a, driver b) {
    drs (a > b) {
        finish a;
    }
    finish b;
}
```

**F1-IR:**
```
define i64 @max(i64 %a, i64 %b) {
entry:
    %0 = gt i64 %a, %b
    br %0, then, else

then:
    ret i64 %a

else:
    ret i64 %b
}
```

---

## 10. Optimizer Passes

### 10.1 Constant Folding

Evaluates constant expressions at compile time.

**Before:**
```
%0 = add i64 3, 5
%1 = mul i64 %0, 2
```

**After:**
```
%0 = copy i64 8
%1 = copy i64 16
```

**Rules:**
- `const op const` → `const`
- `x + 0` → `x`
- `x * 1` → `x`
- `x * 0` → `0`
- `x - x` → `0`

### 10.2 Constant Propagation

Replaces variables with known constant values.

**Before:**
```
%x = copy i64 10
%y = add i64 %x, 5
```

**After:**
```
%x = copy i64 10
%y = add i64 10, 5    // then constant folding makes it 15
```

### 10.3 Dead Code Elimination (DCE)

Removes instructions whose results are never used.

**Before:**
```
%0 = add i64 1, 2     // unused
%1 = mul i64 3, 4
ret i64 %1
```

**After:**
```
%1 = mul i64 3, 4
ret i64 %1
```

**Rules:**
- Remove instructions with unused results
- Remove unreachable basic blocks
- Remove stores to unused variables

### 10.4 Tail Call Optimization Pass

Detects and marks tail calls for optimization.

**Before:**
```
define i64 @fib(i64 %n, i64 %a, i64 %b) {
entry:
    %cond = eq i64 %n, 0
    br %cond, base, recurse

base:
    ret i64 %a

recurse:
    %n1 = sub i64 %n, 1
    %sum = add i64 %a, %b
    %result = call i64 @fib(%n1, %b, %sum)
    ret i64 %result
}
```

**After:**
```
define i64 @fib(i64 %n, i64 %a, i64 %b) {
entry:
    %cond = eq i64 %n, 0
    br %cond, base, recurse

base:
    ret i64 %a

recurse:
    %n1 = sub i64 %n, 1
    %sum = add i64 %a, %b
    tailcall i64 @fib(%n1, %b, %sum)  // converted to tailcall
}
```

**Detection algorithm:**
1. Find all `ret` terminators
2. Check if the returned value comes from a call in the same block
3. Verify no instructions between call and ret use the result
4. If call target is same function, replace with `tailcall` terminator

### 10.5 Pass Ordering

Passes should be applied in this order:
1. **Constant Propagation** — propagate constants first
2. **Constant Folding** — then fold them
3. **Dead Code Elimination** — remove dead code
4. **Tail Call Optimization** — detect tail calls last

Passes may iterate multiple times until no changes occur (fixed-point).

### 10.6 Future Passes (Not Implemented)

- **Common Subexpression Elimination (CSE)**
- **Loop Invariant Code Motion (LICM)**
- **Strength Reduction** (replace expensive ops with cheaper ones)
- **Inlining** (inline small functions)

---

## 11. LLVM Code Generation

### 11.1 Type Mapping

| F1-Lang Type | LLVM Type |
|--------------|-----------|
| `int` | `i64` |
| `bool` | `i1` |
| `string` | `ptr` (pointer to i8 array) |
| `void` | `void` |
| `function` | `ptr` (pointer to closure struct: `{ptr fn, ptr env}`) |

### 11.2 String Handling

Strings are global constants:

```llvm
@.str.0 = private unnamed_addr constant [14 x i8] c"Hello, World!\00"
```

### 11.3 Built-in Function Mapping

```llvm
; External declarations
declare i32 @printf(ptr, ...)
declare i32 @fprintf(ptr, ptr, ...)
declare i32 @strcmp(ptr, ptr)
declare ptr @stderr
declare ptr @malloc(i64)

; Format strings
@.fmt.int = private constant [4 x i8] c"%ld\0A\00"
@.fmt.str = private constant [4 x i8] c"%s\0A\00"
@.str.true = private constant [5 x i8] c"true\00"
@.str.false = private constant [6 x i8] c"false\00"
@.bono.prefix = private constant [28 x i8] c"Bono, my tyres are gone! %s\0A\00"

; radio(int) implementation
call i32 @printf(ptr @.fmt.int, i64 %val)

; radio(string) implementation
call i32 @printf(ptr @.fmt.str, ptr %str)

; radio(bool) implementation
%str = select i1 %boolval, ptr @.str.true, ptr @.str.false
call i32 @printf(ptr @.fmt.str, ptr %str)

; bono(string) implementation
%stderr = load ptr, ptr @stderr
call i32 (ptr, ptr, ...) @fprintf(ptr %stderr, ptr @.bono.prefix, ptr %msg)
```

### 11.4 Graphics Runtime

Graphics functions are implemented via a minimal runtime library linked with the compiled program.

**Runtime globals:**
```llvm
@canvas.buffer = internal global ptr null      ; pixel buffer (RGB bytes)
@canvas.width = internal global i64 0
@canvas.height = internal global i64 0
```

**Runtime functions (implemented in C, linked at compile time):**

```c
// runtime.c - linked with compiled F1-Lang programs

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

static uint8_t* buffer = NULL;
static int64_t width = 0;
static int64_t height = 0;

void f1_canvas(int64_t w, int64_t h) {
    if (buffer) free(buffer);
    width = (w > 4096) ? 4096 : (w < 1 ? 1 : w);
    height = (h > 4096) ? 4096 : (h < 1 ? 1 : h);
    buffer = calloc(width * height * 3, 1);
}

void f1_pixel(int64_t x, int64_t y, int64_t r, int64_t g, int64_t b) {
    if (!buffer) { fprintf(stderr, "error: canvas not initialized\n"); exit(1); }
    if (x < 0 || x >= width || y < 0 || y >= height) return;
    r = (r < 0) ? 0 : (r > 255 ? 255 : r);
    g = (g < 0) ? 0 : (g > 255 ? 255 : g);
    b = (b < 0) ? 0 : (b > 255 ? 255 : b);
    size_t idx = (y * width + x) * 3;
    buffer[idx] = r; buffer[idx+1] = g; buffer[idx+2] = b;
}

void f1_render(const char* filename) {
    if (!buffer) { fprintf(stderr, "error: canvas not initialized\n"); exit(1); }
    FILE* f = fopen(filename, "wb");
    fprintf(f, "P6\n%lld %lld\n255\n", width, height);
    fwrite(buffer, 1, width * height * 3, f);
    fclose(f);
}

void f1_snapshot(int64_t frame) {
    char filename[32];
    snprintf(filename, sizeof(filename), "frame_%04lld.ppm", frame);
    f1_render(filename);
}
```

**LLVM declarations for runtime:**
```llvm
declare void @f1_canvas(i64, i64)
declare void @f1_pixel(i64, i64, i64, i64, i64)
declare void @f1_render(ptr)
declare void @f1_snapshot(i64)
```

**Compilation command:**
```bash
clang -c runtime.c -o runtime.o
clang output.ll runtime.o -o program
```

### 11.5 String Equality

String comparison uses `strcmp`:

```llvm
; a == b (strings)
%cmp = call i32 @strcmp(ptr %a, ptr %b)
%result = icmp eq i32 %cmp, 0

; a != b (strings)
%cmp = call i32 @strcmp(ptr %a, ptr %b)
%result = icmp ne i32 %cmp, 0
```

### 11.6 Control Flow Translation

**If-Else:**
```llvm
  %cond = icmp sgt i64 %a, %b
  br i1 %cond, label %then, label %else

then:
  ; then body
  br label %endif

else:
  ; else body
  br label %endif

endif:
  ; continue
```

**Loop:**
```llvm
  br label %loop.cond

loop.cond:
  %i = phi i64 [0, %entry], [%i.next, %loop.body]
  %cond = icmp slt i64 %i, 10
  br i1 %cond, label %loop.body, label %loop.end

loop.body:
  ; body
  %i.next = add i64 %i, 1
  br label %loop.cond

loop.end:
  ; continue
```

### 11.7 Closure Code Generation

**Environment allocation:**
```llvm
; Allocate environment for 2 captured i64 values
%env = call ptr @malloc(i64 16)
store i64 %x, ptr %env                          ; store first capture
%env.1 = getelementptr i64, ptr %env, i64 1
store i64 %y, ptr %env.1                        ; store second capture
```

**Closure struct:**
```llvm
; Closure = { function_ptr, env_ptr }
%closure = call ptr @malloc(i64 16)
store ptr @lambda.0, ptr %closure               ; function pointer
%closure.env = getelementptr ptr, ptr %closure, i64 1
store ptr %env, ptr %closure.env                ; environment pointer
```

**Lambda function (with hidden env parameter):**
```llvm
; Lambda: pitstop(driver z) { finish x + y + z; }
; Where x, y are captured
define i64 @lambda.0(ptr %env, i64 %z) {
entry:
    %x = load i64, ptr %env
    %env.1 = getelementptr i64, ptr %env, i64 1
    %y = load i64, ptr %env.1
    %sum1 = add i64 %x, %y
    %sum2 = add i64 %sum1, %z
    ret i64 %sum2
}
```

**Calling a closure:**
```llvm
; result = f(arg) where f is a closure
%fn_ptr = load ptr, ptr %f                      ; extract function pointer
%env_slot = getelementptr ptr, ptr %f, i64 1
%env_ptr = load ptr, ptr %env_slot              ; extract environment
%result = call i64 %fn_ptr(ptr %env_ptr, i64 %arg)  ; call with env as first arg
```

### 11.8 Tail Call Code Generation

**Tail call attribute:**
```llvm
; Regular call
%result = call i64 @fib(i64 %n1, i64 %b, i64 %sum)
ret i64 %result

; Optimized to tail call
%result = tail call i64 @fib(i64 %n1, i64 %b, i64 %sum)
ret i64 %result
```

**For guaranteed optimization, use musttail:**
```llvm
; musttail guarantees tail call optimization
%result = musttail call i64 @fib(i64 %n1, i64 %b, i64 %sum)
ret i64 %result
```

**Requirements for musttail:**
- Caller and callee must have same calling convention
- Return types must match
- Must be immediately followed by `ret`

**Generated assembly (x86-64):**
```asm
; Without TCO:         ; With TCO:
call fib               jmp fib        ; no stack growth
ret
```

### 11.9 Complete Example

**Source:**
```
pitstop fib(driver n) {
    drs (n < 2) {
        finish n;
    }
    finish fib(n - 1) + fib(n - 2);
}

radio(fib(10));
```

**LLVM IR:**
```llvm
@.fmt.int = private constant [4 x i8] c"%ld\0A\00"

declare i32 @printf(ptr, ...)

define i64 @fib(i64 %n) {
entry:
  %0 = icmp slt i64 %n, 2
  br i1 %0, label %then, label %else

then:
  ret i64 %n

else:
  %1 = sub i64 %n, 1
  %2 = call i64 @fib(i64 %1)
  %3 = sub i64 %n, 2
  %4 = call i64 @fib(i64 %3)
  %5 = add i64 %2, %4
  ret i64 %5
}

define i32 @main() {
entry:
  %0 = call i64 @fib(i64 10)
  call i32 (ptr, ...) @printf(ptr @.fmt.int, i64 %0)
  ret i32 0
}
```

---

## 12. Example Programs

### 12.1 Hello World

```
radio("Lights out and away we go!");
```

### 12.2 Fibonacci

```
pitstop fib(driver n) {
    drs (n < 2) {
        finish n;
    }
    finish fib(n - 1) + fib(n - 2);
}

radio(fib(20));  // 6765
```

### 12.3 FizzBuzz

```
lap (driver i = 1; i <= 100; i = i + 1) {
    drs (i % 15 == 0) {
        radio("FizzBuzz");
    } defend drs (i % 3 == 0) {
        radio("Fizz");
    } defend drs (i % 5 == 0) {
        radio("Buzz");
    } defend {
        radio(i);
    }
}
```

### 12.4 Factorial

```
pitstop factorial(driver n) {
    drs (n <= 1) {
        finish 1;
    }
    finish n * factorial(n - 1);
}

radio(factorial(10));  // 3628800
```

### 12.5 GCD (Euclidean Algorithm)

```
pitstop gcd(driver a, driver b) {
    drs (b == 0) {
        finish a;
    }
    finish gcd(b, a % b);
}

radio(gcd(48, 18));  // 6
```

### 12.6 Tail-Recursive Fibonacci

```
pitstop fib(driver n, driver a, driver b) {
    drs (n == 0) {
        finish a;
    }
    finish fib(n - 1, b, a + b);  // tail call - no stack growth
}

radio(fib(50, 0, 1));  // 12586269025 (instant, no overflow)
```

### 12.7 Higher-Order Functions

```
// Map a function over a range
pitstop forEach(driver fn, driver start, driver end) {
    lap (driver i = start; i < end; i = i + 1) {
        fn(i);
    }
}

driver printSquare = pitstop(driver x) {
    radio(x * x);
};

forEach(printSquare, 1, 6);  // prints 1, 4, 9, 16, 25
```

### 12.8 Closure Capturing

```
pitstop makeAdder(driver n) {
    finish pitstop(driver x) {
        finish x + n;  // captures 'n'
    };
}

driver add5 = makeAdder(5);
driver add10 = makeAdder(10);

radio(add5(3));   // 8
radio(add10(3));  // 13
```

### 12.9 Function Composition

```
pitstop compose(driver f, driver g) {
    finish pitstop(driver x) {
        finish f(g(x));
    };
}

driver double = pitstop(driver x) { finish x * 2; };
driver addOne = pitstop(driver x) { finish x + 1; };

driver doubleThenAdd = compose(addOne, double);
radio(doubleThenAdd(5));  // 11 (5*2 + 1)
```

---

## 13. CLI Interface

### 13.1 Commands

```bash
# Compile and run immediately
f1c run <file.f1>

# Compile to binary
f1c build <file.f1> [-o output]

# Emit LLVM IR (for debugging/learning)
f1c emit <file.f1>

# Emit F1-IR (our intermediate representation)
f1c emit-ir <file.f1>

# Type check only
f1c check <file.f1>

# Tokenize only (debugging)
f1c lex <file.f1>

# Parse only (debugging)
f1c parse <file.f1>
```

### 13.2 Flags

```bash
-o <file>       # Output file name
-O0             # No optimization (default)
-O1             # Basic optimizations
-v              # Verbose output
--emit-llvm     # Keep .ll file after build
```

---

## 14. Error Messages

### 14.1 Lexer Errors

```
error[L001]: unexpected character '$'
 --> example.f1:3:10
  |
3 | driver x$y = 5;
  |          ^ unexpected character
```

```
error[L002]: invalid escape sequence '\q'
 --> example.f1:2:15
  |
2 | driver s = "hello\qworld";
  |                  ^^ invalid escape sequence
```

```
error[L003]: integer literal overflow
 --> example.f1:1:12
  |
1 | driver x = 99999999999999999999999;
  |            ^^^^^^^^^^^^^^^^^^^^^^^ value exceeds 64-bit signed integer range
```

```
error[L004]: unterminated string literal
 --> example.f1:2:12
  |
2 | driver s = "hello
  |            ^ string literal not closed
```

### 14.2 Parser Errors

```
error[P001]: expected ';' after statement
 --> example.f1:5:1
  |
4 | driver x = 5
5 | driver y = 10;
  | ^ expected ';' here
```

### 14.3 Type Errors

```
error[T001]: type mismatch in binary expression
 --> example.f1:3:14
  |
3 | driver x = 5 + "hello";
  |              ^ cannot add 'int' and 'string'
```

```
error[T002]: undefined variable 'foo'
 --> example.f1:7:8
  |
7 | radio(foo);
  |        ^^^ not found in this scope
```

---

## 15. File Extension

F1-Lang source files use the `.f1` extension.

```
hello.f1
fibonacci.f1
race_simulation.f1
```

---

## Appendix A: Reserved for Future

- Arrays: `driver arr = [1, 2, 3];`
- Structs: `team { driver name; driver points; }`
- Garbage collection for closures (currently environments leak)
- Import system
- Pattern matching
