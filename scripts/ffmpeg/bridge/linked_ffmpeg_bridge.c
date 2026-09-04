#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#include <fcntl.h>
#include <io.h>
#include <windows.h>
#define PODHNOLOGIC_STDOUT_FILENO _fileno(stdout)
#define PODHNOLOGIC_STDERR_FILENO _fileno(stderr)
#else
#include <unistd.h>
#define PODHNOLOGIC_STDOUT_FILENO STDOUT_FILENO
#define PODHNOLOGIC_STDERR_FILENO STDERR_FILENO
#endif

extern int podhnologic_ffmpeg_main(int argc, char **argv);
extern int podhnologic_ffprobe_main(int argc, char **argv);

/* Convert a duplicate of a borrowed Win32 HANDLE to an owned CRT fd. */
#ifdef _WIN32
static int podhnologic_crt_fd_from_handle(uintptr_t value, int flags)
{
    HANDLE source = (HANDLE)(uintptr_t)value;
    HANDLE duplicate = INVALID_HANDLE_VALUE;
    int fd;

    if (!DuplicateHandle(GetCurrentProcess(), source, GetCurrentProcess(),
                         &duplicate, 0, FALSE, DUPLICATE_SAME_ACCESS))
        return -1;

    fd = _open_osfhandle((intptr_t)duplicate, flags);
    if (fd < 0)
        CloseHandle(duplicate);
    return fd;
}
#endif

int podhnologic_linked_ffmpeg_main(const char *tool, int argc, const char **argv, uintptr_t stdout_handle, uintptr_t stderr_handle)
{
    int saved_stdout = -1;
    int saved_stderr = -1;
    int exit_code = 1;
    char **tool_argv = NULL;

    if (!tool)
        return 1;

    tool_argv = calloc((size_t)argc + 2, sizeof(char *));
    if (!tool_argv)
        return 1;

    tool_argv[0] = (char *)tool;
    for (int i = 0; i < argc; i++)
        tool_argv[i + 1] = (char *)argv[i];

#ifdef _WIN32
    saved_stdout = _dup(PODHNOLOGIC_STDOUT_FILENO);
    saved_stderr = _dup(PODHNOLOGIC_STDERR_FILENO);
#else
    saved_stdout = dup(PODHNOLOGIC_STDOUT_FILENO);
    saved_stderr = dup(PODHNOLOGIC_STDERR_FILENO);
#endif
    if (saved_stdout < 0 || saved_stderr < 0)
        goto finish;

#ifdef _WIN32
    {
        int stdout_fd = podhnologic_crt_fd_from_handle(stdout_handle, _O_WRONLY | _O_BINARY);
        if (stdout_fd < 0)
            goto finish;
        if (_dup2(stdout_fd, PODHNOLOGIC_STDOUT_FILENO) < 0) {
            _close(stdout_fd);
            goto finish;
        }
        _close(stdout_fd);
    }
    {
        int stderr_fd = podhnologic_crt_fd_from_handle(stderr_handle, _O_WRONLY | _O_BINARY);
        if (stderr_fd < 0)
            goto finish;
        if (_dup2(stderr_fd, PODHNOLOGIC_STDERR_FILENO) < 0) {
            _close(stderr_fd);
            goto finish;
        }
        _close(stderr_fd);
    }
#else
    if (dup2((int)stdout_handle, PODHNOLOGIC_STDOUT_FILENO) < 0)
        goto finish;
    if (dup2((int)stderr_handle, PODHNOLOGIC_STDERR_FILENO) < 0)
        goto finish;
#endif

    if (strcmp(tool, "ffmpeg") == 0)
        exit_code = podhnologic_ffmpeg_main(argc + 1, tool_argv);
    else if (strcmp(tool, "ffprobe") == 0)
        exit_code = podhnologic_ffprobe_main(argc + 1, tool_argv);

    fflush(stdout);
    fflush(stderr);

finish:
    if (saved_stdout >= 0) {
#ifdef _WIN32
        _dup2(saved_stdout, PODHNOLOGIC_STDOUT_FILENO);
        _close(saved_stdout);
#else
        dup2(saved_stdout, PODHNOLOGIC_STDOUT_FILENO);
        close(saved_stdout);
#endif
    }
    if (saved_stderr >= 0) {
#ifdef _WIN32
        _dup2(saved_stderr, PODHNOLOGIC_STDERR_FILENO);
        _close(saved_stderr);
#else
        dup2(saved_stderr, PODHNOLOGIC_STDERR_FILENO);
        close(saved_stderr);
#endif
    }
    free(tool_argv);
    return exit_code;
}
