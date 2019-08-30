import {
    ServerMessage, Card, RoomResponse, GameEvent, DealResponse,
} from './model';

/**
 * Interface for sending messages and receiving game notifications.
 */
export default interface GameEventHub {
    /**
     * Requests the server to create a new room for the player.
     *
     * @param req Room creation message.
     */
    createRoom(playerId: string, roomName: string, numPlayers: number): void;

    /**
     * Requests the server to join the player in some existing room.
     *
     * @param req Join message.
     */
    joinRoom(playerId: string, roomName: string): void;

    /**
     * Submits the player's card for that turn.
     *
     * @param req Turn message.
     */
    showCard(playerId: string, roomName: string, card: Card): void;

    requestNewGmae(playerId: string, roomName: string): void;

    /**
     * Submits the player's message to other players.
     *
     * @param req Chat message.
     */
    sendMsg(playerId: string, roomName: string, msg: string): void;

    /**
     * Adds a listener for errors.
     *
     * @param callback Callback function.
     * @param persist Whether to persist that callback or destroy it after the first call.
     */
    onError(callback: (msg: string, event: GameEvent) => void, persist?: boolean): void;

    /**
     * Adds a listener for player joining event.
     *
     * @param callback Callback function
     * @param persist Whether to persist that callback or destroy it after the first call.
     */
    onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void, persist?: boolean): void;

    /**
     * Adds a listener for player turn event.
     *
     * @param callback Callback function
     * @param persist Whether to persist that callback or destroy it after the first call.
     */
    onPlayerTurn(callback: (resp: ServerMessage<DealResponse>) => void, persist?: boolean): void;

    /**
     * Adds a listener for player messaging event.
     *
     * @param callback Callback function
     * @param persist Whether to persist that callback or destroy it after the first call.
     */
    onPlayerMsg(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;

    /**
     * Adds a listener for player winning event.
     *
     * @param callback Callback function
     * @param persist Whether to persist that callback or destroy it after the first call.
     */
    onPlayerWin(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;

    /**
     * Adds a listener for game ending.
     *
     * @param callback Callback function
     * @param persist Whether to persist that callback or destroy it after the first call.
     */
    onGameOver(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;

    onGameRequest(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;

    onGameRestart(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;
}
