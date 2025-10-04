'use client';

import { useEffect, useRef, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { WebRTCManager } from '@/lib/webrtc/WebRTCManager';

interface RemotePeer {
  id: string;
  stream: MediaStream;
}

export default function CallRoomPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params.roomId as string;

  const [isConnected, setIsConnected] = useState(false);
  const [isAudioEnabled, setIsAudioEnabled] = useState(true);
  const [isVideoEnabled, setIsVideoEnabled] = useState(true);
  const [isScreenSharing, setIsScreenSharing] = useState(false);
  const [isRecording, setIsRecording] = useState(false);
  const [remotePeers, setRemotePeers] = useState<RemotePeer[]>([]);
  const [error, setError] = useState<string | null>(null);

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const webrtcManagerRef = useRef<WebRTCManager | null>(null);
  const clientIdRef = useRef<string>(`client-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const recordedChunksRef = useRef<Blob[]>([]);
  const recordingStartTimeRef = useRef<number>(0);

  useEffect(() => {
    initializeCall();

    return () => {
      // クリーンアップ
      if (webrtcManagerRef.current) {
        webrtcManagerRef.current.disconnect();
      }
    };
  }, []);

  /**
   * 通話を初期化
   */
  const initializeCall = async () => {
    try {
      // トークンを取得（認証実装に応じて変更）
      const token = localStorage.getItem('accessToken');
      if (!token) {
        router.push('/login');
        return;
      }

      // WebRTCマネージャーを初期化
      const manager = new WebRTCManager(roomId, clientIdRef.current);
      webrtcManagerRef.current = manager;

      // イベントハンドラーを設定
      manager.onRemoteStream = (peerId, stream) => {
        console.log('Remote stream received from:', peerId);
        setRemotePeers(prev => {
          // 既存のピアを更新または新規追加
          const existingIndex = prev.findIndex(p => p.id === peerId);
          if (existingIndex >= 0) {
            const updated = [...prev];
            updated[existingIndex] = { id: peerId, stream };
            return updated;
          }
          return [...prev, { id: peerId, stream }];
        });
      };

      manager.onPeerLeft = (peerId) => {
        console.log('Peer left:', peerId);
        setRemotePeers(prev => prev.filter(p => p.id !== peerId));
      };

      manager.onError = (err) => {
        console.error('WebRTC error:', err);
        setError(err.message);
      };

      // シグナリングサーバーに接続
      await manager.connect(token);

      // ローカルメディアストリームを取得
      const localStream = await manager.getLocalMediaStream({
        audio: true,
        video: true
      });

      // ローカルビデオを表示
      if (localVideoRef.current) {
        localVideoRef.current.srcObject = localStream;
      }

      setIsConnected(true);
    } catch (err) {
      console.error('Failed to initialize call:', err);
      setError(err instanceof Error ? err.message : 'Failed to initialize call');
    }
  };

  /**
   * 音声のミュート/ミュート解除
   */
  const toggleAudio = () => {
    if (webrtcManagerRef.current) {
      const newState = !isAudioEnabled;
      webrtcManagerRef.current.toggleAudio(newState);
      setIsAudioEnabled(newState);
    }
  };

  /**
   * ビデオのオン/オフ
   */
  const toggleVideo = () => {
    if (webrtcManagerRef.current) {
      const newState = !isVideoEnabled;
      webrtcManagerRef.current.toggleVideo(newState);
      setIsVideoEnabled(newState);
    }
  };

  /**
   * 画面共有の開始/停止
   */
  const toggleScreenShare = async () => {
    if (!webrtcManagerRef.current) return;

    try {
      if (isScreenSharing) {
        webrtcManagerRef.current.stopScreenShare();
        setIsScreenSharing(false);
      } else {
        await webrtcManagerRef.current.startScreenShare();
        setIsScreenSharing(true);
      }
    } catch (err) {
      console.error('Failed to toggle screen share:', err);
      setError(err instanceof Error ? err.message : 'Failed to toggle screen share');
    }
  };

  /**
   * 録音の開始/停止
   */
  const toggleRecording = async () => {
    if (!webrtcManagerRef.current) return;

    try {
      if (isRecording) {
        // 録音停止
        if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
          mediaRecorderRef.current.stop();
        }
      } else {
        // 録音開始
        const localStream = webrtcManagerRef.current.getLocalStream();
        if (!localStream) {
          throw new Error('No local stream available');
        }

        // 音声トラックのみを使用
        const audioStream = new MediaStream(localStream.getAudioTracks());

        const mediaRecorder = new MediaRecorder(audioStream, {
          mimeType: 'audio/webm;codecs=opus'
        });

        recordedChunksRef.current = [];
        recordingStartTimeRef.current = Date.now();

        mediaRecorder.ondataavailable = (event) => {
          if (event.data.size > 0) {
            recordedChunksRef.current.push(event.data);
          }
        };

        mediaRecorder.onstop = async () => {
          const blob = new Blob(recordedChunksRef.current, { type: 'audio/webm' });
          const duration = Math.floor((Date.now() - recordingStartTimeRef.current) / 1000);

          // バックエンドに録音データをアップロード
          await uploadRecording(blob, duration);

          setIsRecording(false);
          mediaRecorderRef.current = null;
        };

        mediaRecorder.start(1000); // 1秒ごとにデータを記録
        mediaRecorderRef.current = mediaRecorder;
        setIsRecording(true);
      }
    } catch (err) {
      console.error('Failed to toggle recording:', err);
      setError(err instanceof Error ? err.message : 'Failed to toggle recording');
    }
  };

  /**
   * 録音データをバックエンドにアップロード
   */
  const uploadRecording = async (blob: Blob, duration: number) => {
    try {
      const token = localStorage.getItem('accessToken');
      const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

      const formData = new FormData();
      formData.append('file', blob, `recording-${roomId}-${Date.now()}.webm`);
      formData.append('duration', duration.toString());

      const response = await fetch(`${API_BASE_URL}/api/calls/rooms/${roomId}/recordings`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to upload recording');
      }

      console.log('Recording uploaded successfully');
    } catch (err) {
      console.error('Failed to upload recording:', err);
      setError(err instanceof Error ? err.message : 'Failed to upload recording');
    }
  };

  /**
   * 通話を終了
   */
  const leaveCall = () => {
    // 録音中の場合は停止
    if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
      mediaRecorderRef.current.stop();
    }

    if (webrtcManagerRef.current) {
      webrtcManagerRef.current.disconnect();
    }
    router.push('/calls');
  };

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900">
        <div className="bg-red-500 text-white p-6 rounded-lg max-w-md">
          <h2 className="text-xl font-bold mb-2">Error</h2>
          <p>{error}</p>
          <button
            onClick={() => router.push('/calls')}
            className="mt-4 bg-white text-red-500 px-4 py-2 rounded hover:bg-gray-100"
          >
            Back to Calls
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      {/* ヘッダー */}
      <div className="bg-gray-800 p-4 border-b border-gray-700">
        <div className="container mx-auto flex justify-between items-center">
          <h1 className="text-xl font-bold">Room: {roomId}</h1>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-400">
              {isConnected ? '●' : '○'} {remotePeers.length} participant(s)
            </span>
          </div>
        </div>
      </div>

      {/* ビデオグリッド */}
      <div className="container mx-auto p-4">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* ローカルビデオ */}
          <div className="relative bg-gray-800 rounded-lg overflow-hidden aspect-video">
            <video
              ref={localVideoRef}
              autoPlay
              playsInline
              muted
              className="w-full h-full object-cover"
            />
            <div className="absolute bottom-2 left-2 bg-black bg-opacity-50 px-2 py-1 rounded text-sm">
              You {isScreenSharing && '(Screen)'}
            </div>
            {!isVideoEnabled && (
              <div className="absolute inset-0 flex items-center justify-center bg-gray-700">
                <div className="text-6xl">👤</div>
              </div>
            )}
          </div>

          {/* リモートビデオ */}
          {remotePeers.map((peer) => (
            <RemoteVideo key={peer.id} peer={peer} />
          ))}
        </div>
      </div>

      {/* コントロールバー */}
      <div className="fixed bottom-0 left-0 right-0 bg-gray-800 border-t border-gray-700 p-4">
        <div className="container mx-auto flex justify-center items-center gap-4">
          {/* マイクボタン */}
          <button
            onClick={toggleAudio}
            className={`p-4 rounded-full ${
              isAudioEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-500 hover:bg-red-600'
            }`}
            title={isAudioEnabled ? 'Mute' : 'Unmute'}
          >
            {isAudioEnabled ? '🎤' : '🔇'}
          </button>

          {/* カメラボタン */}
          <button
            onClick={toggleVideo}
            className={`p-4 rounded-full ${
              isVideoEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-500 hover:bg-red-600'
            }`}
            title={isVideoEnabled ? 'Turn off camera' : 'Turn on camera'}
          >
            {isVideoEnabled ? '📹' : '📷'}
          </button>

          {/* 画面共有ボタン */}
          <button
            onClick={toggleScreenShare}
            className={`p-4 rounded-full ${
              isScreenSharing ? 'bg-blue-500 hover:bg-blue-600' : 'bg-gray-700 hover:bg-gray-600'
            }`}
            title={isScreenSharing ? 'Stop sharing' : 'Share screen'}
          >
            🖥️
          </button>

          {/* 録音ボタン */}
          <button
            onClick={toggleRecording}
            className={`p-4 rounded-full ${
              isRecording ? 'bg-red-600 hover:bg-red-700 animate-pulse' : 'bg-gray-700 hover:bg-gray-600'
            }`}
            title={isRecording ? 'Stop recording' : 'Start recording'}
          >
            {isRecording ? '⏹️' : '⏺️'}
          </button>

          {/* 通話終了ボタン */}
          <button
            onClick={leaveCall}
            className="p-4 rounded-full bg-red-500 hover:bg-red-600"
            title="Leave call"
          >
            📞
          </button>
        </div>
      </div>
    </div>
  );
}

/**
 * リモートビデオコンポーネント
 */
function RemoteVideo({ peer }: { peer: RemotePeer }) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (videoRef.current && peer.stream) {
      videoRef.current.srcObject = peer.stream;
    }
  }, [peer.stream]);

  return (
    <div className="relative bg-gray-800 rounded-lg overflow-hidden aspect-video">
      <video
        ref={videoRef}
        autoPlay
        playsInline
        className="w-full h-full object-cover"
      />
      <div className="absolute bottom-2 left-2 bg-black bg-opacity-50 px-2 py-1 rounded text-sm">
        Participant {peer.id.substring(0, 8)}
      </div>
    </div>
  );
}
