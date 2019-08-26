import {
  ClientMessage, RoomCreationRequest, ServerMessage,
  RoomResponse, GameEvent, DealResponse, TurnRequest,
} from './model';
import { Callback } from './connection';

import GameEventHub from './';

export default class TutorialProvider implements GameEventHub {

  private static callbacks: { [key in GameEvent]?: Array<Callback<(resp: any) => void>> } = {};

  private static errorCallbacks: Array<Callback<(msg: string, event: GameEvent) => void>> = [];

  public createRoom(req: ClientMessage<RoomCreationRequest>) {
    console.warn(`Ignoring room creation request for tutorial:`, req);
  }

  public joinRoom(req: ClientMessage<{}>) {
    console.warn(`Ignoring room joining request for tutorial:`, req);
  }

  public showCard(req: ClientMessage<TurnRequest>) {
    //
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
    TutorialProvider.errorCallbacks.push({
      callback,
      persist,
    });
  }

  public propagateMessage(data: ServerMessage<any>) {
    const callbacks = TutorialProvider.callbacks[data.event];
    if (callbacks) {
      callbacks.forEach((c) => {
        c.callback(data);
      });

      TutorialProvider.callbacks[data.event] = callbacks.filter((c) => c.persist);
    } else if (data.msg !== '' && TutorialProvider.errorCallbacks) {
      TutorialProvider.errorCallbacks.forEach((c) => {
        c.callback(data.msg, data.event);
      });

      TutorialProvider.errorCallbacks = TutorialProvider.errorCallbacks.filter((c) => c.persist);
    } else {
      console.warn('Ignoring unknown event response', data);
    }
  }

  private onEvent<R>(event: GameEvent, callback: (resp: ServerMessage<R>) => void, persist?: boolean) {
    if (TutorialProvider.callbacks[event] === undefined) {
      TutorialProvider.callbacks[event] = [];
    }

    TutorialProvider.callbacks[event]!.push({
      callback,
      persist,
    });
  }
}
