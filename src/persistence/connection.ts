import {
  ClientMessage, RoomCreationRequest, ServerMessage,
  RoomResponse, GameEvent, DealResponse, TurnRequest,
} from './model';

import GameEventHub from './';

interface Callback<F> {
  persist: boolean | undefined;
  callback: F;
}

/**
 * Provides a wrapper for `Websocket` object and exposes the messaging interface.
 *
 * This uses statics and is global for the app i.e., only one websocket is created
 * and used throughout the app. Same applies to added event listeners.
 */
export default class ConnectionProvider implements GameEventHub {

  private static conn: WebSocket | null = null;

  private static callbacks: { [key in GameEvent]?: Array<Callback<(resp: any) => void>> } = {};

  private static errorCallbacks: Array<Callback<(msg: string, event: GameEvent) => void>> = [];

  /* Interface methods */

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

  /**
   * Generic event listener for all game events.
   *
   * @param event Game event.
   * @param callback Callback to be called with server response.
   * @param persist Whether to persist that callback or destroy it after the first call.
   */
  private onEvent<R>(event: GameEvent, callback: (resp: ServerMessage<R>) => void, persist?: boolean) {
    if (ConnectionProvider.callbacks[event] === undefined) {
      ConnectionProvider.callbacks[event] = [];
    }

    ConnectionProvider.callbacks[event]!.push({
      callback,
      persist,
    });
  }

  /**
   * Generic message listener for an open websocket.
   *
   * @param event Websocket message event.
   */
  private onMessage(event: MessageEvent) {
    const data: ServerMessage<any> = JSON.parse(event.data);
    console.debug('Incoming message', JSON.stringify(data));
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

  /**
   * Generic sender for messages.
   *
   * @param msg JSON object.
   */
  private sendMessage<T>(msg: T) {
    this.withConnection((ws) => {
      ws.send(JSON.stringify(msg));
    });
  }

  /**
   * Provides the `WebSocket` object to the caller through a callback.
   * Instantiates or reuses the websocket as required.
   */
  private withConnection(callback: (ws: WebSocket) => void) {
    if (ConnectionProvider.conn) {
      return callback(ConnectionProvider.conn);
    }

    let protocol = 'wss';
    if (window.location.protocol.indexOf('https') < 0) {
      protocol = 'ws';
    }

    let url = `${protocol}://${window.location.host}${window.location.pathname}`;
    if (!url.endsWith('/')) {
      url += '/';
    }
    url += 'ws';

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
