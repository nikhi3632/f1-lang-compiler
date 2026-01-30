// F1-Lang Runtime
// Provides built-in functions for F1-Lang programs

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

// ============================================================================
// I/O Functions
// ============================================================================

// radio(int) - prints integer to stdout with newline
void f1_radio_int(int64_t value) {
    printf("%lld\n", (long long)value);
}

// radio(string) - prints string to stdout with newline
void f1_radio_str(const char* str) {
    printf("%s\n", str);
}

// radio(bool) - prints bool to stdout with newline
void f1_radio_bool(int64_t value) {
    printf("%s\n", value ? "true" : "false");
}

// bono(string) - prints error to stderr with F1 meme prefix
void f1_bono(const char* message) {
    fprintf(stderr, "Bono, my tyres are gone! %s\n", message);
}

// ============================================================================
// Graphics Functions
// ============================================================================

static uint8_t* g_buffer = NULL;
static int64_t g_width = 0;
static int64_t g_height = 0;
static char g_frame_dir[256] = "";  // Output directory for snapshot()

// canvas(width, height) - initialize pixel buffer
void f1_canvas(int64_t width, int64_t height) {
    if (g_buffer) {
        free(g_buffer);
    }
    // Clamp dimensions
    g_width = (width < 1) ? 1 : (width > 4096 ? 4096 : width);
    g_height = (height < 1) ? 1 : (height > 4096 ? 4096 : height);
    g_buffer = (uint8_t*)calloc(g_width * g_height * 3, 1);
}

// pixel(x, y, r, g, b) - set pixel color
void f1_pixel(int64_t x, int64_t y, int64_t r, int64_t g, int64_t b) {
    if (!g_buffer) {
        fprintf(stderr, "error: canvas not initialized\n");
        exit(1);
    }
    // Ignore out-of-bounds pixels
    if (x < 0 || x >= g_width || y < 0 || y >= g_height) {
        return;
    }
    // Clamp color values
    r = (r < 0) ? 0 : (r > 255 ? 255 : r);
    g = (g < 0) ? 0 : (g > 255 ? 255 : g);
    b = (b < 0) ? 0 : (b > 255 ? 255 : b);

    size_t idx = (y * g_width + x) * 3;
    g_buffer[idx] = (uint8_t)r;
    g_buffer[idx + 1] = (uint8_t)g;
    g_buffer[idx + 2] = (uint8_t)b;
}

// render(filename) - write canvas to PPM file
void f1_render(const char* filename) {
    if (!g_buffer) {
        fprintf(stderr, "error: canvas not initialized\n");
        exit(1);
    }
    FILE* f = fopen(filename, "wb");
    if (!f) {
        fprintf(stderr, "error: cannot open file %s\n", filename);
        exit(1);
    }
    fprintf(f, "P6\n%lld %lld\n255\n", (long long)g_width, (long long)g_height);
    fwrite(g_buffer, 1, g_width * g_height * 3, f);
    fclose(f);
}

// framedir(path) - set output directory for snapshot()
void f1_framedir(const char* dir) {
    if (dir == NULL || dir[0] == '\0') {
        g_frame_dir[0] = '\0';
    } else {
        snprintf(g_frame_dir, sizeof(g_frame_dir), "%s", dir);
        // Ensure trailing slash
        size_t len = strlen(g_frame_dir);
        if (len > 0 && len < sizeof(g_frame_dir) - 1 && g_frame_dir[len-1] != '/') {
            g_frame_dir[len] = '/';
            g_frame_dir[len+1] = '\0';
        }
    }
}

// snapshot(frame_number) - write canvas to numbered PPM file
void f1_snapshot(int64_t frame) {
    char filename[320];
    snprintf(filename, sizeof(filename), "%sframe_%04lld.ppm", g_frame_dir, (long long)frame);
    f1_render(filename);
}

// ============================================================================
// Utility Functions
// ============================================================================

// Exit with a status code
void f1_exit(int64_t code) {
    exit((int)code);
}
