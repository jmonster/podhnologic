# syntax=docker/dockerfile-upstream:master-labs

FROM emscripten/emsdk:3.1.40 AS emsdk-base
ARG EXTRA_CFLAGS
ARG EXTRA_LDFLAGS
ARG FFMPEG_ST
ARG FFMPEG_MT
ARG FFMPEG_GIT_REF
ENV INSTALL_DIR=/opt
ENV FFMPEG_VERSION=$FFMPEG_GIT_REF
ENV CFLAGS="-I$INSTALL_DIR/include $CFLAGS $EXTRA_CFLAGS"
ENV CXXFLAGS="$CFLAGS"
ENV LDFLAGS="-L$INSTALL_DIR/lib $LDFLAGS $CFLAGS $EXTRA_LDFLAGS"
ENV EM_PKG_CONFIG_PATH=$EM_PKG_CONFIG_PATH:$INSTALL_DIR/lib/pkgconfig:/emsdk/upstream/emscripten/system/lib/pkgconfig
ENV EM_TOOLCHAIN_FILE=$EMSDK/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:$EM_PKG_CONFIG_PATH
ENV FFMPEG_ST=$FFMPEG_ST
ENV FFMPEG_MT=$FFMPEG_MT
RUN apt-get update && \
      apt-get install -y pkg-config autoconf automake libtool ragel curl ca-certificates

FROM emsdk-base AS lame-builder
ARG LAME_TARBALL_URL
ARG LAME_TARBALL_SHA256
RUN curl -fsSL "$LAME_TARBALL_URL" -o /tmp/lame.tar.gz && \
      echo "$LAME_TARBALL_SHA256  /tmp/lame.tar.gz" | sha256sum -c - && \
      mkdir -p /src && \
      tar -xzf /tmp/lame.tar.gz -C /src --strip-components=1
COPY build/lame.sh /src/build.sh
RUN sed -i 's/make install -j/make install -j2/g; s/make -j/make -j2/g' /src/build.sh && bash -x /src/build.sh

FROM emsdk-base AS zlib-builder
ARG ZLIB_TARBALL_URL
ARG ZLIB_TARBALL_SHA256
RUN curl -fsSL "$ZLIB_TARBALL_URL" -o /tmp/zlib.tar.gz && \
      echo "$ZLIB_TARBALL_SHA256  /tmp/zlib.tar.gz" | sha256sum -c - && \
      mkdir -p /src && \
      tar -xzf /tmp/zlib.tar.gz -C /src --strip-components=1
COPY build/zlib.sh /src/build.sh
RUN sed -i 's/make install -j/make install -j2/g; s/make -j/make -j2/g' /src/build.sh && bash -x /src/build.sh

FROM emsdk-base AS ffmpeg-base
ADD https://github.com/FFmpeg/FFmpeg.git#$FFMPEG_VERSION /src
COPY --from=lame-builder $INSTALL_DIR $INSTALL_DIR
COPY --from=zlib-builder $INSTALL_DIR $INSTALL_DIR

FROM ffmpeg-base AS ffmpeg-builder
COPY build/ffmpeg.sh /src/build.sh
ENV AUDIO_DEMUXERS aa,aac,aax,ac3,aiff,ape,asf,au,caf,dsf,dts,eac3,flac,hca,matroska,mov,mp3,mpc,mpc8,ogg,oma,shorten,tak,tta,voc,w64,wav,wv,xwma
ENV AUDIO_DECODERS aac,aac_latm,ac3,alac,ape,atrac1,atrac3,atrac3al,atrac3p,atrac3pal,atrac9,cook,dca,dsd_lsbf,dsd_lsbf_planar,dsd_msbf,dsd_msbf_planar,eac3,flac,hca,mace3,mace6,metasound,mp1,mp1float,mp2,mp2float,mp3,mp3adu,mp3adufloat,mp3float,mp3on4,mp3on4float,mpc7,mpc8,opus,qdm2,qdmc,ra_144,ra_288,ralf,shorten,tak,tta,vorbis,wavpack,wmalossless,wmapro,wmav1,wmav2,bmp,mjpeg,png,adpcm_4xm,adpcm_adx,adpcm_afc,adpcm_agm,adpcm_aica,adpcm_argo,adpcm_circus,adpcm_ct,adpcm_dtk,adpcm_ea,adpcm_ea_maxis_xa,adpcm_ea_r1,adpcm_ea_r2,adpcm_ea_r3,adpcm_ea_xas,adpcm_g722,adpcm_g726,adpcm_g726le,adpcm_ima_acorn,adpcm_ima_alp,adpcm_ima_amv,adpcm_ima_apc,adpcm_ima_apm,adpcm_ima_cunning,adpcm_ima_dat4,adpcm_ima_dk3,adpcm_ima_dk4,adpcm_ima_ea_eacs,adpcm_ima_ea_sead,adpcm_ima_escape,adpcm_ima_hvqm2,adpcm_ima_hvqm4,adpcm_ima_iss,adpcm_ima_magix,adpcm_ima_moflex,adpcm_ima_mtf,adpcm_ima_oki,adpcm_ima_pda,adpcm_ima_qt,adpcm_ima_qt_at,adpcm_ima_rad,adpcm_ima_smjpeg,adpcm_ima_ssi,adpcm_ima_wav,adpcm_ima_ws,adpcm_ima_xbox,adpcm_ms,adpcm_mtaf,adpcm_n64,adpcm_psx,adpcm_psxc,adpcm_sanyo,adpcm_sbpro_2,adpcm_sbpro_3,adpcm_sbpro_4,adpcm_swf,adpcm_thp,adpcm_thp_le,adpcm_vima,adpcm_xa,adpcm_xmd,adpcm_yamaha,adpcm_zork,pcm_alaw,pcm_bluray,pcm_dvd,pcm_f16le,pcm_f24le,pcm_f32be,pcm_f32le,pcm_f64be,pcm_f64le,pcm_lxf,pcm_mulaw,pcm_s16be,pcm_s16be_planar,pcm_s16le,pcm_s16le_planar,pcm_s24be,pcm_s24daud,pcm_s24le,pcm_s24le_planar,pcm_s32be,pcm_s32le,pcm_s32le_planar,pcm_s64be,pcm_s64le,pcm_s8,pcm_s8_planar,pcm_sga,pcm_u16be,pcm_u16le,pcm_u24be,pcm_u24le,pcm_u32be,pcm_u32le,pcm_u8,pcm_vidc
ENV AUDIO_ENCODERS aac,alac,flac,libmp3lame,opus,pcm_alaw,pcm_mulaw,pcm_s16be,pcm_s16le,pcm_s24be,pcm_s24le,pcm_s32be,pcm_s32le,pcm_f32be,pcm_f32le,pcm_u8
ENV AUDIO_PARSERS aac,aac_latm,ac3,cook,dca,flac,mpegaudio,opus,tak,vorbis
RUN sed -i 's/emmake make -j/emmake make -j2/g' /src/build.sh && bash -x /src/build.sh \
      --disable-network \
      --disable-avdevice \
      --enable-small \
      --disable-everything \
      --enable-swresample \
      --enable-protocol=file,pipe \
      --enable-demuxer=$AUDIO_DEMUXERS \
      --enable-muxer=adts,flac,ipod,mov,mp3,ogg,opus,mp4,wav \
      --enable-decoder=$AUDIO_DECODERS \
      --enable-encoder=$AUDIO_ENCODERS \
      --enable-parser=$AUDIO_PARSERS \
      --enable-bsf=aac_adtstoasc \
      --enable-filter=aformat,anull,aresample,atrim \
      --enable-libmp3lame \
      --enable-zlib

FROM ffmpeg-builder AS ffmpeg-wasm-builder
COPY src/bind /src/src/bind
RUN sed -i 's#Module\\[\"_ffmpeg\"\\](args.length, stringsToPtr(args));#Module["ret"] = Module["_ffmpeg"](args.length, stringsToPtr(args));#' /src/src/bind/ffmpeg/bind.js && \
      cp -R /src/fftools /src/src/fftools8 && \
      sed -i 's#int main(int argc, char \\*\\*argv)#EMSCRIPTEN_KEEPALIVE int ffmpeg(int argc, char **argv)#' /src/src/fftools8/ffmpeg.c && \
      for kind in css html; do \
        gzip -9 -c "/src/src/fftools8/resources/graph.$kind" > "/tmp/graph.$kind.gz"; \
        { \
          printf '#include <stdint.h>\n'; \
          printf 'const unsigned char ff_graph_%s_data[] = {\n' "$kind"; \
          od -An -v -t u1 "/tmp/graph.$kind.gz" | tr -s ' ' '\n' | sed '/^$/d; s/$/,/'; \
          printf '};\n'; \
          printf 'const unsigned int ff_graph_%s_len = sizeof(ff_graph_%s_data);\n' "$kind" "$kind"; \
        } > "/src/src/fftools8/resources/graph.$kind.c"; \
      done
COPY build/podhnologic-ffmpeg-wasm.sh build.sh
ENV FFMPEG_LIBS \
      -lmp3lame \
      -lz
RUN mkdir -p /src/dist/umd && bash -x /src/build.sh \
      ${FFMPEG_LIBS} \
      -o dist/umd/ffmpeg-core.js
RUN mkdir -p /src/dist/esm && bash -x /src/build.sh \
      ${FFMPEG_LIBS} \
      -sEXPORT_ES6 \
      -o dist/esm/ffmpeg-core.js

FROM scratch AS exportor
COPY --from=ffmpeg-wasm-builder /src/dist /dist
