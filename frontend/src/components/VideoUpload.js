import React, { useState, useEffect } from 'react';
import axios from 'axios';

const VideoUpload = () => {
    const [selectedFile, setSelectedFile] = useState(null);
    const [uploadStatus, setUploadStatus] = useState('');
    const [processingStatus, setProcessingStatus] = useState('');
    const [elapsedTime, setElapsedTime] = useState(0);
    const [finalElapsedTime, setFinalElapsedTime] = useState(null); // new state variable
    const [intervalId, setIntervalId] = useState(null);
    const [dots, setDots] = useState('');

    useEffect(() => {
        let dotsInterval;
        if (processingStatus.startsWith('Current status: transcribing audio')) {
            dotsInterval = setInterval(() => {
                setDots(prevDots => {
                    if (prevDots.length >= 3) return '';
                    return prevDots + '.';
                });
            }, 1000);
        }

        return () => {
            if (dotsInterval) {
                clearInterval(dotsInterval);
            }
        };
    }, [processingStatus]);

    const handleFileChange = (event) => {
        setSelectedFile(event.target.files[0]);
        setUploadStatus('');
        setProcessingStatus('');
        setElapsedTime(0);
        setFinalElapsedTime(null); // reset final elapsed time
        setDots('');
        if (intervalId) {
            clearInterval(intervalId);
            setIntervalId(null);
        }
    };

    const handleSubmit = async (event) => {
        event.preventDefault();
        if (!selectedFile) {
            setUploadStatus('Please select a file.');
            return;
        }

        const formData = new FormData();
        formData.append('video', selectedFile);

        setUploadStatus('Uploading file...');

        try {
            const response = await axios.post('http://localhost:8080/upload', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                },
                onUploadProgress: (progressEvent) => {
                    let percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
                    setUploadStatus(`Upload progress: ${percentCompleted}%`);
                }
            });

            setUploadStatus('Upload complete.');
            setProcessingStatus('Current status: transcribing audio');

            // Start the timer and dot animation
            const startTime = Date.now();
            const id = setInterval(() => {
                setElapsedTime(Math.round((Date.now() - startTime) / 1000));
            }, 1000);
            setIntervalId(id);

            const checkStatus = async () => {
                const processingStatusResponse = await axios.get('http://localhost:8080/status');
                return processingStatusResponse.data;
            };

            // Continuously check the processing status
            let status = await checkStatus();
            while (status !== 'done') {
                setProcessingStatus(`Current status: ${status}`);
                await new Promise(resolve => setTimeout(resolve, 2000)); // wait for 2 seconds before checking again
                status = await checkStatus();
            }

            clearInterval(id);
            const finalTime = Math.round((Date.now() - startTime) / 1000);
            setFinalElapsedTime(finalTime); // store final elapsed time
            setProcessingStatus(`Transcription complete (${formatTime(finalTime)})`);
        } catch (error) {
            setUploadStatus(`Upload failed: ${error.message}`);
        }
    };

    const formatTime = (seconds) => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins}m ${secs}s`;
    };

    return (
        <div>
            <h2>Upload Video</h2>
            <form onSubmit={handleSubmit}>
                <input type="file" onChange={handleFileChange} />
                <button type="submit">Upload</button>
            </form>
            <p>{uploadStatus}</p>
            <p>{processingStatus}{dots}</p>
            {processingStatus && finalElapsedTime === null && <p>Elapsed time: {formatTime(elapsedTime)}</p>}
            {finalElapsedTime !== null && <p>Elapsed time: {formatTime(finalElapsedTime)}</p>}
        </div>
    );
};

export default VideoUpload;
