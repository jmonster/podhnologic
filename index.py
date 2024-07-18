#!/usr/bin/env python3

import os
import sys
import json
import subprocess
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor

def parse_arguments():
    import argparse
    parser = argparse.ArgumentParser(description="Convert audio files using ffmpeg.")
    parser.add_argument('--input', required=True, help='Input directory')
    parser.add_argument('--output', required=True, help='Output directory')
    parser.add_argument('--codec', choices=['flac', 'alac', 'aac', 'wav', 'mp3', 'opus'], help='Audio codec')
    parser.add_argument('--ffmpeg', default='ffmpeg', help='Path to ffmpeg executable')
    parser.add_argument('--dry-run', action='store_true', help='Dry run mode')
    parser.add_argument('--ipod', action='store_true', help='iPod compatibility mode')
    args = parser.parse_args()

    if not args.codec and not args.ipod:
        parser.error('The --codec argument is required unless --ipod is specified.')

    # If --ipod is specified and no codec is provided, default to 'aac'
    if args.ipod and not args.codec:
        args.codec = 'aac'

    return args

def is_audio_file(file_path):
    audio_extensions = {'.mp3', '.wav', '.flac', '.aac', '.opus', '.m4a'}
    return file_path.suffix.lower() in audio_extensions

def ffmpeg_has_aac_at(ffmpeg_path):
    try:
        subprocess.run([ffmpeg_path, "-h", "encoder=aac_at"], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return "aac_at"
    except subprocess.CalledProcessError:
        return "aac"

def escape_shell_arg(arg):
    return f'"{arg}"'

def extract_metadata(ffmpeg_path, file_path):
    command = [ffmpeg_path.replace('ffmpeg', 'ffprobe'), '-v', 'quiet', '-print_format', 'json', '-show_format', '-show_streams', str(file_path)]
    result = subprocess.run(command, capture_output=True, text=True)
    return json.loads(result.stdout)

def get_codec_params(codec, metadata, ipod, ffmpeg_path):
    normalized_tags = {k.lower(): v for k, v in metadata.get('format', {}).get('tags', {}).items()}
    desired_metadata_keys = ['title', 'artist', 'album', 'date', 'track', 'genre', 'disc']
    desired_metadata = ' '.join(
        f'-metadata {key}={escape_shell_arg(normalized_tags[key])}' for key in desired_metadata_keys if key in normalized_tags
    )

    base_params = f"-map 0 -map_metadata -1 {desired_metadata}"
    video_params = "-c:v copy"

    codec_params = {
        'alac': f"-c:a alac {video_params} {'-sample_fmt s16p -ar 44100 -movflags +faststart -disposition:a 0' if ipod else ''}",
        'flac': f"-c:a flac {video_params}",
        'wav': "-c:a pcm_s16le -vn",
        'opus': "-c:a libopus -b:a 128k -vn",
        'aac': f"-c:a {ffmpeg_has_aac_at(ffmpeg_path)} -b:a 256k {video_params}",
        'mp3': "-c:a libmp3lame -q:a 0",
    }

    return f"{base_params} {codec_params[codec]}"

def convert_file(ffmpeg_path, input_file_path, output_file_path, codec_params, dry_run):
    command = f'{ffmpeg_path} -i {escape_shell_arg(str(input_file_path))} {codec_params} {escape_shell_arg(str(output_file_path))} > /dev/null 2>&1'

    if dry_run:
        print(f"[dry run] Command: {command}")
    else:
        if output_file_path.exists():
            print(f"File exists, skipping: {output_file_path}")
            return
        output_file_path.parent.mkdir(parents=True, exist_ok=True)
        print(f"Converting {input_file_path} to {output_file_path}")
        try:
            subprocess.run(command, shell=True, check=True)
        except subprocess.CalledProcessError as e:
            print(f"Error converting file {input_file_path}: {e}")

def process_files(input_dir, output_dir, codec, ffmpeg_path, dry_run, ipod):
    codec_extension = {'alac': '.m4a', 'flac': '.flac', 'wav': '.wav', 'opus': '.opus', 'aac': '.m4a', 'mp3': '.mp3'}
    with ThreadPoolExecutor(max_workers=os.cpu_count()) as executor:
        futures = []
        for file_path in Path(input_dir).rglob('*'):
            if is_audio_file(file_path):
                metadata = extract_metadata(ffmpeg_path, file_path)
                relative_path = file_path.relative_to(input_dir)
                output_file_path = Path(output_dir) / relative_path
                output_file_path = output_file_path.with_suffix(codec_extension[codec])
                codec_params = get_codec_params(codec, metadata, ipod, ffmpeg_path)
                futures.append(executor.submit(convert_file, ffmpeg_path, file_path, output_file_path, codec_params, dry_run))

        for future in futures:
            try:
                future.result()
            except Exception as e:
                print(f"Error processing file: {e}")

def main():
    args = parse_arguments()
    process_files(args.input, args.output, args.codec, args.ffmpeg, args.dry_run, args.ipod)

if __name__ == '__main__':
    main()
