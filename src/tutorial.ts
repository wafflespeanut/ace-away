import {
  Card, ServerMessage, RoomResponse, GameEvent, DealResponse,
} from './persistence/model';
import { Callback } from './persistence/connection';

import GameEventHub from './persistence';

export default class TutorialProvider implements GameEventHub {

  private callbacks: { [key in GameEvent]?: Array<Callback<(resp: any) => void>> } = {};

  private errorCallbacks: Array<Callback<(msg: string, event: GameEvent) => void>> = [];

  public createRoom(playerId: string, roomName: string, numPlayers: number) {
    console.warn(`Ignoring room creation request for tutorial.`);
  }

  public joinRoom(playerId: string, roomName: string) {
    console.warn(`Ignoring room joining request for tutorial.`);
  }

  public sendMsg(playerId: string, roomName: string, msg: string) {
    console.warn(`Ignoring player message during tutorial.`);
  }

  public showCard(playerId: string, roomName: string, card: Card) {
    //
  }

  public onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerJoin, callback, persist);
  }

  public onPlayerTurn(callback: (resp: ServerMessage<DealResponse>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerTurn, callback, persist);
  }

  public onPlayerMsg(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    console.warn(`Ignoring message response for tutorial.`);
  }

  public onPlayerWin(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    this.onEvent(GameEvent.playerWins, callback, persist);
  }

  public onGameOver(callback: (resp: ServerMessage<{}>) => void, persist?: boolean) {
    //
  }

  public onError(callback: (msg: string, event: GameEvent) => void, persist?: boolean) {
    this.errorCallbacks.push({
      callback,
      persist,
    });
  }

  public propagateMessage(data: ServerMessage<any>) {
    const callbacks = this.callbacks[data.event];
    if (callbacks) {
      callbacks.forEach((c) => {
        c.callback(data);
      });

      this.callbacks[data.event] = callbacks.filter((c) => c.persist);
    } else if (data.msg !== '' && this.errorCallbacks) {
      this.errorCallbacks.forEach((c) => {
        c.callback(data.msg, data.event);
      });

      this.errorCallbacks = this.errorCallbacks.filter((c) => c.persist);
    } else {
      console.warn('Ignoring unknown event response', data);
    }
  }

  private onEvent<R>(event: GameEvent, callback: (resp: ServerMessage<R>) => void, persist?: boolean) {
    if (this.callbacks[event] === undefined) {
      this.callbacks[event] = [];
    }

    this.callbacks[event]!.push({
      callback,
      persist,
    });
  }
}
