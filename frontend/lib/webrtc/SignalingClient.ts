/**
 * WebSocketシグナリングクライアント
 * バックエンドのWebSocketサーバーと通信してWebRTCシグナリングを処理
 */

export interface SignalingMessage {
  type: 'offer' | 'answer' | 'ice-candidate' | 'join' | 'leave' | 'user-joined' | 'user-left';
  from?: string;
  to?: string;
  data?: any;
}

export class SignalingClient {
  private ws: WebSocket | null = null;
  private roomId: string;
  private clientId: string;
  private onMessageCallback: ((message: SignalingMessage) => void) | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(roomId: string, clientId: string) {
    this.roomId = roomId;
    this.clientId = clientId;
  }

  /**
   * WebSocket接続を確立
   */
  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080';
      const url = `${wsUrl}/ws/signaling/${this.roomId}?token=${encodeURIComponent(token)}`;

      // 接続タイムアウト (10秒)
      const timeout = setTimeout(() => {
        if (this.ws && this.ws.readyState === WebSocket.CONNECTING) {
          this.ws.close();
          reject(new Error('WebSocket connection timeout'));
        }
      }, 10000);

      this.ws = new WebSocket(url);

      // 認証トークンをヘッダーに追加できないため、接続後に送信
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        clearTimeout(timeout);
        this.reconnectAttempts = 0;

        // 認証情報を送信
        this.send({
          type: 'join',
          data: { token, clientId: this.clientId }
        });

        resolve();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: SignalingMessage = JSON.parse(event.data);
          console.log('Received message:', message);

          if (this.onMessageCallback) {
            this.onMessageCallback(message);
          }
        } catch (error) {
          console.error('Failed to parse message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        clearTimeout(timeout);
        reject(error);
      };

      this.ws.onclose = () => {
        console.log('WebSocket closed');
        clearTimeout(timeout);
        this.handleReconnect();
      };
    });
  }

  /**
   * メッセージを送信
   */
  send(message: SignalingMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        ...message,
        from: this.clientId
      }));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  /**
   * メッセージ受信時のコールバックを設定
   */
  onMessage(callback: (message: SignalingMessage) => void): void {
    this.onMessageCallback = callback;
  }

  /**
   * 接続を切断
   */
  disconnect(): void {
    if (this.ws) {
      this.send({ type: 'leave' });
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * 再接続処理
   */
  private handleReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);

      setTimeout(() => {
        // 再接続にはトークンが必要だが、ここでは取得できないため
        // 上位レイヤーで再接続を処理する必要がある
      }, this.reconnectDelay * this.reconnectAttempts);
    }
  }

  /**
   * 接続状態を取得
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}
