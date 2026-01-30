; F1-Lang LLVM IR
target triple = "x86_64-unknown-linux-gnu"

; External declarations
declare void @f1_radio_int(i64)
declare void @f1_radio_str(ptr)
declare void @f1_radio_bool(i1)
declare void @f1_bono(ptr)
declare void @f1_canvas(i64, i64)
declare void @f1_pixel(i64, i64, i64, i64, i64)
declare void @f1_render(ptr)
declare void @f1_snapshot(i64)
declare i32 @strcmp(ptr, ptr)
declare ptr @malloc(i64)

@str.0 = private constant [27 x i8] c"Lights out and away we go!\00"

define void @main() {
entry.0:
    call void @f1_radio_str(ptr @str.0)
    ret void
}
