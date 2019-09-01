import {
  ClientMessage, ServerMessage,
  RoomResponse, GameEvent, DealResponse, Card,
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

  private static socketErrorCallbacks: Array<Callback<() => void>> = [];

  private static disconnectCallbacks: Array<Callback<() => void>> = [];

  /* Interface methods */

  public createRoom(playerId: string, roomName: string, numPlayers: number) {
    this.sendMessage({
      player: playerId,
      room: roomName,
      event: GameEvent.createRoom,
      data: {
        players: numPlayers,
      },
    });
  }

  public joinRoom(playerId: string, roomName: string) {
    this.sendMessage({
      player: playerId,
      room: roomName,
      event: GameEvent.playerJoin,
      data: {},
    });
  }

  public showCard(playerId: string, roomName: string, card: Card) {
    this.sendMessage({
      player: playerId,
      room: roomName,
      event: GameEvent.playerTurn,
      data: {
        card,
      },
    });
  }

  public requestNewGmae(playerId: string, roomName: string) {
    this.sendMessage({
      player: playerId,
      room: roomName,
      event: GameEvent.restartRequest,
      data: {},
    });
  }

  public sendMsg(playerId: string, roomName: string, msg: string) {
    this.sendMessage({
      player: playerId,
      room: roomName,
      event: GameEvent.playerMsg,
      data: {},
      msg,
    });
  }

  public onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerJoin, callback, persist);
  }

  public onPlayerTurn(callback: (resp: ServerMessage<DealResponse>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerTurn, callback, persist);
  }

  public onPlayerMsg(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerMsg, callback, persist);
  }

  public onPlayerWin(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerWins, callback, persist);
  }

  public onGameOver(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.gameOver, callback, persist);
  }

  public onGameRequest(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.restartRequest, callback, persist);
  }

  public onGameRestart(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.gameRestart, callback, persist);
  }

  public onError(callback: (msg: string, event: GameEvent) => void, persist?: boolean) {
    ConnectionProvider.errorCallbacks.push({
      callback,
      persist,
    });
  }

  public onSocketClose(callback: () => void, persist?: boolean) {
    ConnectionProvider.disconnectCallbacks.push({
      callback,
      persist,
    });
  }

  public onSocketError(callback: () => void, persist?: boolean) {
    ConnectionProvider.socketErrorCallbacks.push({
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
  private sendMessage<T>(msg: ClientMessage<T>) {
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

    socket.onerror = () => {
      ConnectionProvider.socketErrorCallbacks.forEach((c) => c.callback());
      ConnectionProvider.socketErrorCallbacks = ConnectionProvider.socketErrorCallbacks.filter((c) => c.persist);
    };

    socket.onclose = () => {
      ConnectionProvider.disconnectCallbacks.forEach((c) => c.callback());
      ConnectionProvider.disconnectCallbacks = ConnectionProvider.disconnectCallbacks.filter((c) => c.persist);
    };
  }
}

export { Callback };
