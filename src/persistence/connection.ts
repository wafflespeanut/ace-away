import {
  ClientMessage, RoomCreationRequest, ServerMessage,
  RoomResponse, GameEvent, DealResponse, TurnRequest,
} from './model';

import GameEventHub from './';

interface Callback<F> {
  persist: boolean | undefined;
  callback: F;
}

export default class ConnectionProvider implements GameEventHub {

  private static conn: WebSocket | null = null;

  private static callbacks: { [key in GameEvent]?: Array<Callback<(resp: any) => void>> } = {};

  private static errorCallbacks: Array<Callback<(msg: string, event: GameEvent) => void>> = [];

  public createRoom(req: ClientMessage<RoomCreationRequest>) {
    this.sendMessage(req);
  }

  public joinRoom(req: ClientMessage<{}>) {
    this.sendMessage(req);
  }

  public showCard(req: ClientMessage<TurnRequest>) {
    this.sendMessage(req);
  }

  public onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerJoin, callback, persist);
  }

  public onPlayerTurn(callback: (resp: ServerMessage<DealResponse>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerTurn, callback, persist);
  }

  public onPlayerWin(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerWins, callback, persist);
  }

  public onGameOver(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.gameOver, callback, persist);
  }

  public onError(callback: (msg: string, event: GameEvent) => void, persist?: boolean) {
    ConnectionProvider.errorCallbacks.push({
      callback,
      persist,
    });
  }

  private onEvent<R>(event: GameEvent, callback: (resp: ServerMessage<R>) => void, persist?: boolean) {
    if (ConnectionProvider.callbacks[event] === undefined) {
      ConnectionProvider.callbacks[event] = [];
    }

    ConnectionProvider.callbacks[event]!.push({
      callback,
      persist,
    });
  }

  private onMessage(event: MessageEvent) {
    const data: ServerMessage<any> = JSON.parse(event.data);
    console.debug('Incoming message', data);
    const callbacks = ConnectionProvider.callbacks[data.event];
    if (callbacks) {
      callbacks.forEach((c) => {
        c.callback(data);
      });

      ConnectionProvider.callbacks[data.event] = callbacks.filter((c) => c.persist);
    } else if (data.msg !== '' && ConnectionProvider.errorCallbacks) {
      ConnectionProvider.errorCallbacks.forEach((c) => {
        c.callback(data.msg, data.event);
      });

      ConnectionProvider.errorCallbacks = ConnectionProvider.errorCallbacks.filter((c) => c.persist);
    } else {
      console.warn('Ignoring unknown event response', data);
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
