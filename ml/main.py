import sys
import os
import whisper
import json

def transcribe_audio(audio_path):
    model = whisper.load_model("large")
    result = model.transcribe(audio_path, language="pt")
    return result

def save_transcript(transcript, output_dir, base_filename):
    # Save raw transcript
    raw_transcript_path = os.path.join(output_dir, f"{base_filename}.txt")
    with open(raw_transcript_path, 'w', encoding='utf-8') as f:
        f.write(transcript['text'])
    
    # Save transcript with timestamps in SRT format
    srt_transcript_path = os.path.join(output_dir, f"{base_filename}.srt")
    with open(srt_transcript_path, 'w', encoding='utf-8') as f:
        for i, segment in enumerate(transcript['segments']):
            start = segment['start']
            end = segment['end']
            text = segment['text']
            f.write(f"{i + 1}\n")
            f.write(f"{format_time(start)} --> {format_time(end)}\n")
            f.write(f"{text}\n\n")

    # Save transcript as JSON
    json_transcript_path = os.path.join(output_dir, f"{base_filename}.json")
    with open(json_transcript_path, 'w', encoding='utf-8') as f:
        json.dump(transcript, f, ensure_ascii=False, indent=4)

def format_time(seconds):
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    seconds = seconds % 60
    milliseconds = int((seconds % 1) * 1000)
    return f"{hours:02}:{minutes:02}:{int(seconds):02},{milliseconds:03}"


def process_audio(audio_path):
    print(f"Processing audio file: {audio_path}")

    # Transcrever o áudio
    transcript = transcribe_audio(audio_path)
    print(f"Transcript: {transcript['text']}")

    # Save transcripts
    base_filename = os.path.splitext(os.path.basename(audio_path))[0]
    output_dir = os.path.dirname(audio_path)
    save_transcript(transcript, output_dir, base_filename)

    return transcript

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python main.py <audio_path>")
        sys.exit(1)
    
    audio_path = sys.argv[1]
    transcript = process_audio(audio_path)
    print(json.dumps(transcript, ensure_ascii=False, indent=4))
