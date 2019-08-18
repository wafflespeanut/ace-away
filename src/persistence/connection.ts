import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse, GameEvent } from './model';
import GameEventHub from './';

export default class ConnectionProvider implements GameEventHub {

  private static conn: WebSocket | null = null;

  private callbacks: { [key in GameEvent]?: (resp: any) => void } = {};

  public createRoom(req: ClientMessage<RoomCreationRequest>) {
    this.sendMessage(req);
  }

  public onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void) {
    this.callbacks[GameEvent.playerJoin] = callback;
  }

  private onMessage(event: MessageEvent) {
    const data: ServerMessage<any> = JSON.parse(event.data);
    const cb = this.callbacks[data.event];
    if (cb) {
      cb(data);
    } else {
      console.debug('Ignoring unknown event response', data);
    }
  }

  private sendMessage<T>(msg: T) {
    this.withConnection((ws) => {
      ws.send(JSON.stringify(msg));
    });
  }

  private withConnection(callback: (ws: WebSocket) => void) {
    if (ConnectionProvider.conn) {
      return callback(ConnectionProvider.conn);
    }

    let protocol = 'wss';
    if (window.location.protocol.indexOf('https') < 0) {
      protocol = 'ws';
    }

    const url = `${protocol}://${window.location.host}/ws`;
    console.log(`Initiating websocket to ${url}.`);
    const socket = new WebSocket(url);
    socket.onopen = () => {
      ConnectionProvider.conn = socket;
      callback(socket);
    };

    socket.onmessage = (e) => {
      this.onMessage(e);
    };
  }
}
