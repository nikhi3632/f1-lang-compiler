// F1-Lang Runtime
// Provides basic I/O operations for F1-Lang programs

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

// Print an integer without newline
int64_t print(int64_t value) {
    printf("%lld", (long long)value);
    return 0;
}

// Print an integer with newline
int64_t println(int64_t value) {
    printf("%lld\n", (long long)value);
    return 0;
}

// Print a string (null-terminated)
int64_t printstr(const char* str) {
    printf("%s", str);
    return 0;
}

// Print a string with newline
int64_t printstrln(const char* str) {
    printf("%s\n", str);
    return 0;
}

// Read an integer from stdin
int64_t readint(void) {
    int64_t value;
    if (scanf("%lld", (long long*)&value) != 1) {
        return 0;
    }
    return value;
}

// Convert boolean to string representation
const char* boolstr(int64_t value) {
    return value ? "true" : "false";
}

// Print a boolean
int64_t printbool(int64_t value) {
    printf("%s", value ? "true" : "false");
    return 0;
}

// Print a boolean with newline
int64_t printboolln(int64_t value) {
    printf("%s\n", value ? "true" : "false");
    return 0;
}

// Exit with a status code
void f1_exit(int64_t code) {
    exit((int)code);
}
