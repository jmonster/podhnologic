#!/bin/bash

set -euo pipefail

EXPORT_NAME="createFFmpegCore"
FFTOOLS_DIR="src/fftools8"

CONF_FLAGS=(
  -I.
  -Icompat/stdbit
  -I"$FFTOOLS_DIR"
  -I"$INSTALL_DIR/include"
  -L"$INSTALL_DIR/lib"
  -Llibavcodec
  -Llibavfilter
  -Llibavformat
  -Llibavutil
  -Llibswresample
  -Llibswscale
  -lavcodec
  -lavfilter
  -lavformat
  -lavutil
  -lswresample
  -lswscale
  -Wno-deprecated-declarations
  -include
  pthread.h
  -include
  emscripten.h
  $LDFLAGS
  -sENVIRONMENT=worker
  -sWASM_BIGINT
  -sSTACK_SIZE=5MB
  -sMODULARIZE
  ${FFMPEG_MT:+ -sINITIAL_MEMORY=1024MB}
  ${FFMPEG_MT:+ -sPTHREAD_POOL_SIZE=32}
  ${FFMPEG_ST:+ -sINITIAL_MEMORY=32MB -sALLOW_MEMORY_GROWTH}
  -sEXPORT_NAME="$EXPORT_NAME"
  -sEXPORTED_FUNCTIONS=_ffmpeg,_abort,_malloc
  -sEXPORTED_RUNTIME_METHODS=FS,setValue,getValue,UTF8ToString,lengthBytesUTF8,stringToUTF8
  -lworkerfs.js
  --pre-js src/bind/ffmpeg/bind.js
  "$FFTOOLS_DIR/cmdutils.c"
  "$FFTOOLS_DIR/ffmpeg.c"
  "$FFTOOLS_DIR/ffmpeg_dec.c"
  "$FFTOOLS_DIR/ffmpeg_demux.c"
  "$FFTOOLS_DIR/ffmpeg_enc.c"
  "$FFTOOLS_DIR/ffmpeg_filter.c"
  "$FFTOOLS_DIR/ffmpeg_hw.c"
  "$FFTOOLS_DIR/ffmpeg_mux.c"
  "$FFTOOLS_DIR/ffmpeg_mux_init.c"
  "$FFTOOLS_DIR/ffmpeg_opt.c"
  "$FFTOOLS_DIR/ffmpeg_sched.c"
  "$FFTOOLS_DIR/graph/graphprint.c"
  "$FFTOOLS_DIR/opt_common.c"
  "$FFTOOLS_DIR/resources/graph.css.c"
  "$FFTOOLS_DIR/resources/graph.html.c"
  "$FFTOOLS_DIR/resources/resman.c"
  "$FFTOOLS_DIR/sync_queue.c"
  "$FFTOOLS_DIR/textformat/avtextformat.c"
  "$FFTOOLS_DIR/textformat/tf_compact.c"
  "$FFTOOLS_DIR/textformat/tf_default.c"
  "$FFTOOLS_DIR/textformat/tf_flat.c"
  "$FFTOOLS_DIR/textformat/tf_ini.c"
  "$FFTOOLS_DIR/textformat/tf_json.c"
  "$FFTOOLS_DIR/textformat/tf_mermaid.c"
  "$FFTOOLS_DIR/textformat/tf_xml.c"
  "$FFTOOLS_DIR/textformat/tw_avio.c"
  "$FFTOOLS_DIR/textformat/tw_buffer.c"
  "$FFTOOLS_DIR/textformat/tw_stdout.c"
  "$FFTOOLS_DIR/thread_queue.c"
)

emcc "${CONF_FLAGS[@]}" "$@"
