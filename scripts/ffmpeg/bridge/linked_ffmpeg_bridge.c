#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

extern int podhnologic_ffmpeg_main(int argc, char **argv);
extern int podhnologic_ffprobe_main(int argc, char **argv);

int podhnologic_linked_ffmpeg_main(const char *tool, int argc, const char **argv, int stdout_fd, int stderr_fd)
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

    saved_stdout = dup(STDOUT_FILENO);
    saved_stderr = dup(STDERR_FILENO);
    if (saved_stdout < 0 || saved_stderr < 0)
        goto finish;

    if (dup2(stdout_fd, STDOUT_FILENO) < 0)
        goto finish;
    if (dup2(stderr_fd, STDERR_FILENO) < 0)
        goto finish;

    if (strcmp(tool, "ffmpeg") == 0) {
        exit_code = podhnologic_ffmpeg_main(argc + 1, tool_argv);
    } else if (strcmp(tool, "ffprobe") == 0) {
        exit_code = podhnologic_ffprobe_main(argc + 1, tool_argv);
    }

    fflush(stdout);
    fflush(stderr);

finish:
    if (saved_stdout >= 0) {
        dup2(saved_stdout, STDOUT_FILENO);
        close(saved_stdout);
    }
    if (saved_stderr >= 0) {
        dup2(saved_stderr, STDERR_FILENO);
        close(saved_stderr);
    }
    free(tool_argv);
    return exit_code;
}
