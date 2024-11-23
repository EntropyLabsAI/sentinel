import React, { useRef, useEffect, useState } from 'react';
import WaveSurfer from 'wavesurfer.js';
import RegionsPlugin from 'wavesurfer.js/dist/plugins/regions.js';
import TimelinePlugin from 'wavesurfer.js/dist/plugins/timeline.js';

interface AudioBacktestProps {
  onFileUpload?: (file: File) => void;
}

export default function AudioBacktestComponent() {

  return <AudioBacktest onFileUpload={() => { }} />
}

const AudioBacktest: React.FC<AudioBacktestProps> = ({ onFileUpload }) => {
  const waveformRef = useRef<HTMLDivElement>(null);
  const wavesurfer = useRef<WaveSurfer | null>(null);
  const [duration, setDuration] = useState<number>(0);
  const [fileName, setFileName] = useState<string>('');
  const wsRegions = useRef<RegionsPlugin | null>(null);
  const [isPlaying, setIsPlaying] = useState<boolean>(false);
  const [currentTime, setCurrentTime] = useState<number>(0);
  const [errorSegments, setErrorSegments] = useState<{ time: number, error: string }[]>([]);

  const ERROR_TYPES = [
    "Customer couldn't understand agent",
    "Agent talking over customer",
    "Customer insults agent",
    "Agent misunderstood question",
    "Long period of silence",
    "Agent provided incorrect information",
    "Customer speaking too quietly",
    "Multiple people speaking simultaneously",
    "Background noise interference"
  ];

  useEffect(() => {
    if (waveformRef.current) {
      wavesurfer.current = WaveSurfer.create({
        container: waveformRef.current,
        waveColor: '#6366f1', // Indigo-500
        progressColor: '#4338ca', // Indigo-700
        height: 100,
        normalize: true,
        splitChannels: [{ overlay: false }],
        // minPxPerSec: 10,
        fillParent: true,
        // responsive: true,
      });

      wsRegions.current = wavesurfer.current?.registerPlugin(RegionsPlugin.create());

      wavesurfer.current.on('ready', () => {
        if (wavesurfer.current) {
          setDuration(Math.ceil(wavesurfer.current.getDuration()));
        }
      });

      wavesurfer.current.on('play', () => setIsPlaying(true));
      wavesurfer.current.on('pause', () => setIsPlaying(false));
      wavesurfer.current.on('timeupdate', (time) => setCurrentTime(time));

      return () => {
        wavesurfer.current?.destroy();
      };
    }
  }, []);

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file && wavesurfer.current) {
      setFileName(file.name);
      wavesurfer.current.loadBlob(file);
      onFileUpload?.(file);
    }

  };

  useEffect(() => {
    // Set error segments to n random seconds between 0 and duration
    if (duration > 0) {
      const n = 5;
      const randomErrorSegments = Array.from({ length: n }, () => ({
        time: Math.floor(Math.random() * duration),
        error: ERROR_TYPES[Math.floor(Math.random() * ERROR_TYPES.length)]
      }));
      setErrorSegments(randomErrorSegments);
    }
  }, [duration]);

  useEffect(() => {
    if (wavesurfer.current && duration > 0) {
      errorSegments.forEach(({ time }) => {
        wsRegions.current?.addRegion({
          start: time,
          end: time + 1,
          color: 'rgba(239, 68, 68, 0.5)', // Red-500 with opacity
          drag: false,
          resize: false,
        });
      });
    }
  }, [errorSegments, duration]);

  const handlePlayPause = () => {
    wavesurfer.current?.playPause();
  };

  const handleSeekToTime = (second: number) => {
    if (wavesurfer.current) {
      wavesurfer.current.setTime(second);
    }
  };

  return (
    <div className="m-12 p-6 bg-white rounded-md border">
      <div className="mb-6">
        <label className="inline-flex items-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 
                         text-white rounded-lg cursor-pointer transition-colors duration-200 
                         shadow-sm focus-within:ring-2 focus-within:ring-offset-2 
                         focus-within:ring-indigo-500">
          <svg
            className="w-5 h-5 mr-2"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>
          {fileName ? 'Change Audio File' : 'Upload Audio File'}
          <input
            type="file"
            accept="audio/*"
            onChange={handleFileUpload}
            className="hidden"
          />
        </label>
        {fileName && (
          <span className="ml-3 text-sm text-gray-600">
            {fileName}
          </span>
        )}
      </div>

      <div className="bg-gray-50 p-4 rounded-lg">
        <div ref={waveformRef} className="w-full max-w-full overflow-hidden" />

        {fileName && (
          <div className="mt-4 flex items-center gap-4">
            <button
              onClick={handlePlayPause}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg"
            >
              {isPlaying ? 'Pause' : 'Play'}
            </button>
            <span className="text-sm text-gray-600">
              {Math.floor(currentTime)}s / {Math.ceil(duration)}s
            </span>
          </div>
        )}
      </div>

      {errorSegments.length > 0 && (
        <div className="mt-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">
            Detected Errors:
          </h3>
          <div className="flex flex-col gap-2">
            {errorSegments
              .sort((a, b) => a.time - b.time)
              .map(({ time, error }) => (
                <div
                  key={time}
                  onClick={() => handleSeekToTime(time)}
                  className="group flex items-center p-3 bg-red-50 border border-red-200 
                           rounded-lg cursor-pointer hover:bg-red-100 transition-colors"
                >
                  <div className="flex-1">
                    <div className="text-red-800 font-medium">
                      {error}
                    </div>
                    <div className="text-red-600 text-sm">
                      Timestamp: {time}s
                    </div>
                  </div>
                  <div className="text-red-400 group-hover:text-red-600">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                        d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                        d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                </div>
              ))}
          </div>
        </div>
      )}
    </div>
  );
};
