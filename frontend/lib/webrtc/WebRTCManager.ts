/**
 * WebRTC接続を管理するクラス
 * 複数のピア接続を管理し、音声・映像ストリームを処理
 */

import { SignalingClient, SignalingMessage } from './SignalingClient';

export interface MediaStreamConfig {
  audio: boolean;
  video: boolean;
}

export interface PeerConnection {
  peerId: string;
  connection: RTCPeerConnection;
  stream?: MediaStream;
}

export class WebRTCManager {
  private signalingClient: SignalingClient;
  private localStream: MediaStream | null = null;
  private screenStream: MediaStream | null = null;
  private peerConnections: Map<string, RTCPeerConnection> = new Map();
  private iceServers: RTCIceServer[] = [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' }
  ];

  // イベントハンドラー
  onRemoteStream?: (peerId: string, stream: MediaStream) => void;
  onPeerLeft?: (peerId: string) => void;
  onError?: (error: Error) => void;

  constructor(roomId: string, clientId: string) {
    this.signalingClient = new SignalingClient(roomId, clientId);
    this.setupSignalingHandlers();
  }

  /**
   * シグナリングハンドラーの設定
   */
  private setupSignalingHandlers(): void {
    this.signalingClient.onMessage(async (message: SignalingMessage) => {
      try {
        switch (message.type) {
          case 'user-joined':
            if (message.from) {
              await this.handleUserJoined(message.from);
            }
            break;

          case 'offer':
            if (message.from && message.data) {
              await this.handleOffer(message.from, message.data);
            }
            break;

          case 'answer':
            if (message.from && message.data) {
              await this.handleAnswer(message.from, message.data);
            }
            break;

          case 'ice-candidate':
            if (message.from && message.data) {
              await this.handleIceCandidate(message.from, message.data);
            }
            break;

          case 'user-left':
            if (message.from) {
              this.handleUserLeft(message.from);
            }
            break;
        }
      } catch (error) {
        console.error('Error handling signaling message:', error);
        if (this.onError) {
          this.onError(error as Error);
        }
      }
    });
  }

  /**
   * 接続を開始
   */
  async connect(token: string): Promise<void> {
    await this.signalingClient.connect(token);
  }

  /**
   * ローカルメディアストリームを取得
   */
  async getLocalMediaStream(config: MediaStreamConfig = { audio: true, video: true }): Promise<MediaStream> {
    try {
      this.localStream = await navigator.mediaDevices.getUserMedia(config);
      return this.localStream;
    } catch (error) {
      console.error('Failed to get local stream:', error);
      throw error;
    }
  }

  /**
   * 画面共有を開始
   */
  async startScreenShare(): Promise<MediaStream> {
    try {
      this.screenStream = await navigator.mediaDevices.getDisplayMedia({
        video: true,
        audio: false
      });

      // 画面共有が停止されたときの処理
      this.screenStream.getVideoTracks()[0].onended = () => {
        this.stopScreenShare();
      };

      // 既存の接続に画面共有トラックを追加
      this.replaceVideoTrack(this.screenStream.getVideoTracks()[0]);

      return this.screenStream;
    } catch (error) {
      console.error('Failed to start screen share:', error);
      throw error;
    }
  }

  /**
   * 画面共有を停止
   */
  stopScreenShare(): void {
    if (this.screenStream) {
      this.screenStream.getTracks().forEach(track => track.stop());
      this.screenStream = null;

      // カメラに戻す
      if (this.localStream) {
        const videoTrack = this.localStream.getVideoTracks()[0];
        this.replaceVideoTrack(videoTrack);
      }
    }
  }

  /**
   * ビデオトラックを置き換え
   */
  private replaceVideoTrack(newTrack: MediaStreamTrack): void {
    this.peerConnections.forEach(pc => {
      const sender = pc.getSenders().find(s => s.track?.kind === 'video');
      if (sender) {
        sender.replaceTrack(newTrack);
      }
    });
  }

  /**
   * 新しいユーザーが参加したときの処理
   */
  private async handleUserJoined(peerId: string): Promise<void> {
    console.log('User joined:', peerId);

    // Peer Connectionを作成
    const pc = this.createPeerConnection(peerId);

    // Offerを作成して送信
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);

    this.signalingClient.send({
      type: 'offer',
      to: peerId,
      data: offer
    });
  }

  /**
   * Offerを受信したときの処理
   */
  private async handleOffer(peerId: string, offer: RTCSessionDescriptionInit): Promise<void> {
    console.log('Received offer from:', peerId);

    const pc = this.createPeerConnection(peerId);

    await pc.setRemoteDescription(new RTCSessionDescription(offer));

    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);

    this.signalingClient.send({
      type: 'answer',
      to: peerId,
      data: answer
    });
  }

  /**
   * Answerを受信したときの処理
   */
  private async handleAnswer(peerId: string, answer: RTCSessionDescriptionInit): Promise<void> {
    console.log('Received answer from:', peerId);

    const pc = this.peerConnections.get(peerId);
    if (pc) {
      await pc.setRemoteDescription(new RTCSessionDescription(answer));
    }
  }

  /**
   * ICE Candidateを受信したときの処理
   */
  private async handleIceCandidate(peerId: string, candidate: RTCIceCandidateInit): Promise<void> {
    const pc = this.peerConnections.get(peerId);
    if (pc) {
      await pc.addIceCandidate(new RTCIceCandidate(candidate));
    }
  }

  /**
   * ユーザーが退出したときの処理
   */
  private handleUserLeft(peerId: string): void {
    console.log('User left:', peerId);

    const pc = this.peerConnections.get(peerId);
    if (pc) {
      pc.close();
      this.peerConnections.delete(peerId);
    }

    if (this.onPeerLeft) {
      this.onPeerLeft(peerId);
    }
  }

  /**
   * Peer Connectionを作成
   */
  private createPeerConnection(peerId: string): RTCPeerConnection {
    if (this.peerConnections.has(peerId)) {
      return this.peerConnections.get(peerId)!;
    }

    const pc = new RTCPeerConnection({
      iceServers: this.iceServers
    });

    // ローカルストリームを追加
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => {
        pc.addTrack(track, this.localStream!);
      });
    }

    // ICE Candidateの送信
    pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.signalingClient.send({
          type: 'ice-candidate',
          to: peerId,
          data: event.candidate
        });
      }
    };

    // リモートストリームの受信
    pc.ontrack = (event) => {
      console.log('Received remote stream from:', peerId);
      if (this.onRemoteStream && event.streams[0]) {
        this.onRemoteStream(peerId, event.streams[0]);
      }
    };

    // 接続状態の監視
    pc.onconnectionstatechange = () => {
      console.log('Connection state:', pc.connectionState);
      if (pc.connectionState === 'failed' || pc.connectionState === 'closed') {
        this.peerConnections.delete(peerId);
      }
    };

    this.peerConnections.set(peerId, pc);
    return pc;
  }

  /**
   * 音声のミュート/ミュート解除
   */
  toggleAudio(enabled: boolean): void {
    if (this.localStream) {
      this.localStream.getAudioTracks().forEach(track => {
        track.enabled = enabled;
      });
    }
  }

  /**
   * ビデオのオン/オフ
   */
  toggleVideo(enabled: boolean): void {
    if (this.localStream) {
      this.localStream.getVideoTracks().forEach(track => {
        track.enabled = enabled;
      });
    }
  }

  /**
   * 接続を切断
   */
  disconnect(): void {
    // すべてのPeer Connectionをクローズ
    this.peerConnections.forEach(pc => pc.close());
    this.peerConnections.clear();

    // ローカルストリームを停止
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => track.stop());
      this.localStream = null;
    }

    // 画面共有を停止
    if (this.screenStream) {
      this.screenStream.getTracks().forEach(track => track.stop());
      this.screenStream = null;
    }

    // シグナリング接続を切断
    this.signalingClient.disconnect();
  }

  /**
   * ローカルストリームを取得
   */
  getLocalStream(): MediaStream | null {
    return this.localStream;
  }

  /**
   * 接続中のピア数を取得
   */
  getPeerCount(): number {
    return this.peerConnections.size;
  }
}
