<template>
  <v-app>
    <v-navigation-drawer temporary v-model="drawerOpen" :width="drawerWidth" app>
      <v-row class="d-flex flex-column mx-2 my-4 caption">
        <v-card v-for="(msg, i) in messages" :key="i"
                class="my-1"
                :ref="(i === messages.length - 1) ? 'lastMessage' : null">
          <v-row class="px-2">
            <v-col :class="['pb-0', 'body-2', msg.color + '--text']">{{ msg.sender }}</v-col>
            <v-col :class="['pb-0', 'text-right']">{{ msg.time }}</v-col>
          </v-row>
          <v-card-text class="pa-2 caption">{{ msg.content }}</v-card-text>
        </v-card>
      </v-row>
      <template v-slot:append>
        <v-textarea placeholder="Send a message."
                    v-model="playerMsg"
                    @keydown.enter.exact.native.prevent
                    @keyup.enter.exact.native="postMessage"
                    @keydown.enter.shift.exact.native="playerMsg = playerMsg + '\n'"
                    rows="3" solo no-resize>
          <template v-slot:append>
            <v-btn @click="postMessage" icon>
              <v-icon>mdi-comment</v-icon>
            </v-btn>
          </template>
        </v-textarea>
      </template>
    </v-navigation-drawer>
    <v-app-bar app>
      <v-badge color="red"
               v-if="roomJoined && !showTutorial"
               overlap>
        <template v-slot:badge>
          <span v-if="unreadMessages > 0">{{ unreadMessages }}</span>
        </template>
        <v-btn @click.stop="drawerOpen = !drawerOpen; unreadMessages = 0;" icon>
          <v-icon large>mdi-comment</v-icon>
        </v-btn>
      </v-badge>
      <v-toolbar-title></v-toolbar-title>
      <v-row class="d-flex justify-center">
        <v-slide-y-transition>
          <v-alert v-if="alertMsg"
                    border="top" colored-border
                    :type="alertType" elevation="2">{{ alertMsg }}</v-alert>
        </v-slide-y-transition>
      </v-row>
      <v-btn :disabled="roomJoined === null" @click="restartRequest = true" icon>
        <v-icon large>mdi-restart</v-icon>
      </v-btn>
      <v-btn :disabled="true" icon @click="beginTutorial">
        <v-icon>mdi-help-circle-outline</v-icon>
      </v-btn>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-row class="my-5"></v-row>
        <v-row class="d-flex flex-column flex-md-row">
          <v-row v-if="table.length" :class="[{
            'mx-10': $vuetify.breakpoint.mdAndUp
          }, 'mb-10']">
            <v-card :width="tableSize" :height="tableSize" class="mx-auto my-auto"
                    :style="{ borderRadius: '50%' }">
              <transition-group name="fade">
                <v-card :style="tableCardStyles(i)"
                        class="d-flex flex-column align-center justify-center"
                        v-for="(item, i) in table" :key="i + 0"
                        :height="0.9 * cardSize" :width="0.9 * cardSize"
                        :elevation="item.turn ? 8 : 2"
                        :color="winners[item.id] ? 'green darken-3' : (item.turn ? 'cyan darken-3' : '' )">
                  <div v-if="item.card" class="headline">{{ item.card.label }} {{ prettyMap[item.card.suite] }}</div>
                  <div v-else-if="winners[item.id]">[no cards]</div>
                  <div v-else-if="item.turn">???</div>
                  <div v-else>-</div>
                  <div class="caption">{{ item.id }}</div>
                </v-card>
              </transition-group>
              <transition v-if="tableLockTime !== null" name="fade">
                <v-progress-circular :style="progressStyle"
                                     color="red accent-2"
                                     :value="tableLockTime"
                                     :size="progressSize"
                                     :width="progressWidth"></v-progress-circular>
              </transition>
            </v-card>
          </v-row>
          <v-row :style="{
            'flex-grow': $vuetify.breakpoint.mdAndUp ? '0' : '1',
          }" class="d-flex justify-center align-center mt-5 mb-5">
            <v-tooltip v-if="roomJoined && !showTutorial" bottom>
              <template v-slot:activator="{ on }">
                <v-fab-transition>
                  <v-btn fab icon
                         :width="$vuetify.breakpoint.xs ? 80 : 100"
                         :height="$vuetify.breakpoint.xs ? 80 : 100"
                         :color="selectedCard ? (selectedCard.suite === 'h' || selectedCard.suite === 'd') ? 'red' : 'blue' : ''"
                         :disabled="cardIndex === null || tableLockTime !== null"
                         v-on="$vuetify.breakpoint.smAndUp ? on : {}"
                         @click="sendToPile">
                    <v-icon :class="{
                      'display-3': $vuetify.breakpoint.xs,
                      'display-4': $vuetify.breakpoint.smAndUp,
                    }">{{ cardIndex !== null ? iconMap[selectedCard.suite] : 'mdi-cards-playing-outline' }}</v-icon>
                  </v-btn>
                </v-fab-transition>
              </template>
              <span>Add your card to the pile</span>
            </v-tooltip>
          </v-row>
          <v-col :cols="$vuetify.breakpoint.mdAndUp ? '5' : 'auto'" :class="[{
            'mx-5': $vuetify.breakpoint.xs,
          }, {
            'mx-10': $vuetify.breakpoint.smAndUp,
          }, 'my-auto']">
            <!-- We're "conditionally mandating" because we don't need a card selected all the time. -->
            <v-item-group :mandatory="cardIndex !== null" v-model="cardIndex">
              <v-row class="d-flex my-4" v-for="(item, x) in hand" :key="x">
                  <transition-group class="row" name="card-group">
                    <v-item v-for="l in item.labels" :key="l.label + item.suite"
                            v-slot:default="{ active, toggle }">
                      <v-card :color="active ? ( item.suite === 'h' || item.suite === 'd' ? 'red' : 'blue' ) : ''"
                              @click="() => { selectedCard = { suite: item.suite, label: l.label }; toggle(); }"
                              :height="cardSize"
                              :width="cardSize"
                              class="d-flex justify-center align-center">
                        <span class="display-1">{{ l.label }} {{ prettyMap[item.suite] }}</span>
                      </v-card>
                    </v-item>
                  </transition-group>
              </v-row>
            </v-item-group>
          </v-col>
        </v-row>
        <v-overlay opacity="0.5" v-if="overlayMsgs.length > 0">
          <div>
            <span class="display-1">{{ overlayMsgs[0] }}</span>
            <v-btn class="mb-4 ml-3" icon color="blue" @click="overlayMsgs.splice(0, 1)">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </div>
        </v-overlay>
      </v-container>
      <RequestRestart :showDialog="restartRequest"
                      :fromPlayer="restartRequestees ? restartRequestees[restartRequestees.length - 1] : null"
                      @cancel="restartRequest = false"
                      @request="requestGameRestart" />
      <JoinRoom @player-set="v => playerID = v" :players="allowedPlayers" :showDialog="roomJoined === null" />
      <Tutorial @close="endTutorial"
                @tutorial-step="stepTutorial"
                :playerNeedsHelp="showTutorial" />
    </v-content>
    <v-snackbar class="mt-4" v-if="notification !== null" :value="true" :timeout="0">
      {{ notification }}
      <v-btn color="pink" text @click="notification = null">close</v-btn>
    </v-snackbar>
    <v-footer app></v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import JoinRoom from './dialog/JoinRoom.vue';
import RequestRestart from './dialog/RequestRestart.vue';
import Tutorial, { TutorialStep } from './dialog/Tutorial.vue';
import {
  Card, Suite, suitePrettyMap, Label, PlayerCard,
  GameEvent, suiteIndices, labelRanks, DealResponse,
} from './persistence/model';
import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './persistence/model';
import ConnectionProvider from './persistence/connection';
import TutorialProvider from './tutorial';
import GameEventHub from './persistence';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];
const START_ANGLE = Math.PI / 2;
const TABLE_LOCK_TIME_MS = 5000;
const TURN_WAIT_MS = 10000;

// All colors from https://vuetifyjs.com/en/styles/colors#material-colors.
const COLORS = [
  'indigo',
  'pink',
  'light-green',
  'deep-purple',
  'grey',
  'lime',
  'blue',
  'light-blue',
  'blue-grey',
  'brown',
  'cyan',
  'orange',
  'green',
  'amber',
  'yellow',
  'teal',
  'red',
  'purple',
  'deep-orange',
];

const iconMap = {
  h: 'mdi-cards-heart',
  s: 'mdi-cards-spade',
  c: 'mdi-cards-club',
  d: 'mdi-cards-diamond',
};

/** Label for some suite. */
interface CardLabel {
  label: string;
  updated: boolean;
}

/** Suite in hand. Also holds the labels grouped under that suite. */
interface HandItem {
  suite: string;
  labels: CardLabel[];
}

/** Item in table. Holds player and card information. */
interface TableItem {
  id: string;
  card: Card | null;
  turn: boolean;
}

/** Message object containing a message received from the server. */
interface Message {
  sender: string;
  color: string;
  time: string;
  content: string;
}

/**
 * Performs a linear search through the array and finds the index for inserting
 * the new element using the given compare function. If the element already exists,
 * then its index is returned.
 *
 * @param items Array to be searched.
 * @param newItem New element to be inserted.
 * @param compare Compare function that returns a number.
 */
function searchSortedIndex<T>(items: T[], newItem: T, compare: (e1: T, e2: T) => number) {
  if (items.length === 0 || compare(items[0], newItem) >= 0) {
    return 0;
  }

  let i = 1;
  for (i; i < items.length; i++) {
    const first = compare(items[i - 1], newItem);
    if (first === 0) {
      return i - 1;
    }

    const second = compare(items[i], newItem);
    if (second === 0) {
      return i;
    }

    if (first < 0 && second > 0) {
      return i;
    }
  }

  return i;
}

/** Returns the color for a player from known palette using the player ID. */
function getColorForPlayer(id: string): string {
  const sum = id.split('').reduce((old, value) => old + value.charCodeAt(0), 0);
  const idx = sum % COLORS.length;
  return COLORS[idx];
}

@Component({
  components: {
    JoinRoom,
    RequestRestart,
    Tutorial,
  },
})
export default class App extends Vue {

  /* Internal properties */

  private conn: GameEventHub = new ConnectionProvider();

  /** ID of the most recently set timeout for notifying player turn. */
  private turnNotifyTimeoutId: number = -1;

  /** Whether the player has allowed notifications. */
  private canNotify: boolean = false;

  /** Whether the user is focusing on the app. */
  private hasFocus: boolean = false;

  /* Constants used by models */

  /** Allowed choices for players in rooms. */
  private readonly allowedPlayers: number[] = ALLOWED_PLAYERS;

  /** Object for mapping suites to their unicode representations. */
  private readonly prettyMap: any = suitePrettyMap;

  /** Object for mapping suites to their indices in hand. */
  private readonly indexMap: any = suiteIndices;

  /** Object for mapping suites to their MD icons. */
  private readonly iconMap: any = iconMap;

  /* Readonly models (models that shouldn't be updated manually). */

  /** Whether the app drawer is open. */
  private readonly drawerOpen: boolean = false;

  /** Name set by the player (propagated by `JoinRoom`, after creating/joining a room). */
  private readonly playerID: string = '';

  /* Models */

  /** Whether the modal should be shown for player to issue a restart request */
  private restartRequest: boolean = false;

  /** Players who've requested restarts. */
  private restartRequestees: string[] = [];

  /** Whether the player has opened tutorial. */
  private showTutorial: boolean = false;

  /** Name of the room joined by the player (set after creating/joining a room). */
  private roomJoined: string | null = null;

  /** Notifications from the server shown as a snackbar at the bottom. */
  private notification: string | null = null;

  /** Alert message shown in app bar. */
  private alertMsg: string | null = null;

  /** Type of the alert. */
  private alertType: string = 'info';

  /** Message shown in overlay. */
  private overlayMsgs: string[] = [];

  /** Card selected by the player. */
  private selectedCard: Card | null = null;

  /** Index of `selectedCard`. **Only used for resetting selection. Don't use this directly.** */
  private cardIndex: number | null = null;

  /** Player's hand containing their cards sorted by their labels and suites. */
  private hand: HandItem[] = [];

  /** Table containing the cards from all players for that round. */
  private table: TableItem[] = [];

  /** Winners in this room. */
  private winners: { [s: string]: boolean; } = {};

  /** Number of cards in table in the previous round */
  private previousTurnLength: number = 0;

  /** Number indicating whether the table is locked (happens at the end of each round). */
  private tableLockTime: number | null = null;

  /** Message written by the player. */
  private playerMsg: string = '';

  /** Unread messages since this player has opened the drawer. */
  private unreadMessages: number = 0;

  /** Messages received from the server. */
  private messages: Message[] = [];

  /* Style thingies */

  /** Width of nav drawer. */
  private drawerWidth: number = 300;

  /** Size of the card based on viewports. */
  private get cardSize(): number {
    if (screen.width <= 600) {
      return 85;
    } else if (screen.width <= 960) {
      return 95;
    } else {
      return 100;
    }
  }

  /** Size of the table based on viewports. */
  private get tableSize(): number {
    if (screen.width <= 600) {
      return 300;
    } else if (screen.width <= 960) {
      return 400;
    } else {
      return 450;
    }
  }

  /** Progress bar size (px). */
  private progressSize: number = this.tableSize / 2;

  /** Progress bar width (px). */
  private progressWidth: number = 8;

  /** Styles for progress bar. */
  private get progressStyle(): object {
    const transPos = this.tableSize / 2 - this.progressSize / 2 + this.progressWidth / 2;
    return {
      position: 'absolute',
      transform: `translate(${transPos}px, ${transPos}px)`,
    };
  }

  /** Styles applied to the cards within the table. */
  private tableCardStyles(idx: number): object {
    const total = this.table.length ? this.table.length : 1;
    const angle = 2 * Math.PI / total;
    const x = (this.tableSize / 2) + Math.cos(START_ANGLE + idx * angle) * (this.tableSize / 2) - this.cardSize / 2;
    const y = (this.tableSize / 2) + Math.sin(START_ANGLE + idx * angle) * (this.tableSize / 2) - this.cardSize / 2;

    return {
      transition: `all 500ms ease`,
      transform: `translate(${x}px, ${y}px)`,
      position: 'absolute',
      borderRadius: '4px',
    };
  }

  /* Vue methods */

  private created() {
    this.initialize();
    this.prepareForNotifications();
    this.addListeners();
  }

  /* Init helpers */

  /** Initialize this component. Particularly useful to reset this component. */
  private initialize() {
    this.showTutorial = false;
    this.roomJoined = null;
    this.winners = {};
    this.previousTurnLength = 0;
    this.tableLockTime = null;
    this.cardIndex = null;
    this.selectedCard = null;
    this.conn = new ConnectionProvider();

    this.hand = Object.keys(suiteIndices)
      .sort((a, b) => suiteIndices[a] - suiteIndices[b])
      .map((s) => {
        return {
          suite: s,
          labels: [],
        };
      });
  }

  /** Prepare this view for notifying the player. */
  private prepareForNotifications() {
    document.addEventListener('visibilitychange', () => {
      this.hasFocus = !document.hidden;
    });

    if (!('Notification' in window)) {
      console.warn('This browser does not support desktop notifications.');
      return;
    }

    if (Notification.permission !== 'denied') {
      Notification.requestPermission().then((permission) => {
        this.canNotify = Notification.permission === 'granted';
      });
    }
  }

  /* Add event listeners for changing game state on notifications. */
  private addListeners() {
    this.conn.onError(this.showError, true);
    this.conn.onPlayerJoin(this.playerJoined, true);
    this.conn.onPlayerTurn(this.handlePlayerTurn, true);
    this.conn.onPlayerWin(this.playerWon, true);
    this.conn.onGameOver(this.gameEnded, true);
    this.conn.onPlayerMsg(this.addMessage, true);
    this.conn.onGameRestart((resp) => {
      // This will automatically initiate a cooldown for refreshing the table.
      this.previousTurnLength = Number.POSITIVE_INFINITY;
      this.initialize();
    });

    this.conn.onGameRequest((resp) => {
      const idx = this.restartRequestees.findIndex((id) => id === resp.player);
      if (idx >= 0) {
        this.restartRequestees.splice(idx, 1);
      }

      this.restartRequestees.push(resp.player);
      if (this.restartRequestees.findIndex((id) => id === this.playerID) === -1) {
        this.restartRequest = true;
      }
    }, true);
  }

  /* Game events */

  /** Sends the player-selected card to the pile of cards in the table. */
  private sendToPile() {
    console.log(`Player placing ${this.selectedCard!.label}${this.prettyMap[this.selectedCard!.suite]}`);
    this.conn.showCard(this.playerID, this.roomJoined!, this.selectedCard!);
    this.cardIndex = null;
  }

  /** Posts a message to everyone in the room. */
  private postMessage() {
    this.playerMsg = this.playerMsg.trim();
    if (this.playerMsg !== '') {
      this.conn.sendMsg(this.playerID, this.roomJoined!, this.playerMsg);
    }

    this.playerMsg = '';
  }

  /** A player has joined the room. */
  private playerJoined(resp: ServerMessage<RoomResponse>) {
    this.roomJoined = resp.room;

    resp.response.escaped.forEach((id) => {
      this.winners[id] = true;
    });

    // Set player states.
    this.table = resp.response.players.map((id, i) => {
      return {
        id,
        card: null,
        turn: resp.response.turnIdx === i,
      };
    });

    // Notify if we're waiting on player(s) joining.
    const diff = resp.response.max - resp.response.players.length;
    if (diff > 0) {
      this.alertType = 'info';
      this.alertMsg = `Waiting for ${diff} more player(s) in room ${resp.room}.`;
    } else {
      this.showAlert(`Yay! Let's begin!`);
    }
  }

  /** A player has made their turn. Prepare for next turn. */
  private handlePlayerTurn(resp: ServerMessage<DealResponse>) {
    const previousLength = this.previousTurnLength;
    const currentLength = resp.response.table.length;
    this.previousTurnLength = currentLength;
    clearTimeout(this.turnNotifyTimeoutId);

    const updateStuff = () => {
      // Sort the hand based on suites followed by labels.
      this.updateHand(resp.response.hand);

      // Reset states of cards in our table (if the table isn't locked).
      this.table.forEach((v) => {
        v.card = null;
        v.turn = v.id === resp.response.turnPlayer;
      });

      // Get the cards and set them in our table.
      resp.response.table.forEach((c) => {
        const idx = this.table.findIndex((v) => v.id === c.id); // This will exist.
        this.table[idx].card = c.card;
      });

      // If it's player's turn, notify them if they haven't played for a while.
      if (this.canNotify && resp.response.turnPlayer === this.playerID) {
        this.turnNotifyTimeoutId = setTimeout(() => {
          if (this.table.findIndex((v) => v.id === this.playerID && v.turn) !== -1) {
            const n = new Notification(`Your turn!`, {
              body: `Player(s) are waiting on your turn.`,
            });
          }
        }, TURN_WAIT_MS);
      }
    };

    if (currentLength < previousLength) {
      // If the table is getting cleared, lock the table and
      // pause for a moment for players to see what happened.
      // We're fine delaying this turn because we don't allow the
      // players to place a card in this interval, and so we won't
      // get any `playerTurn` events.
      const timeout = this.initiateTableLockdown();
      setTimeout(() => {
        this.showAlert('Table cleared');
        updateStuff();
      }, timeout);
    } else {
      updateStuff();
    }
  }

  /** Adds incoming message from the server and notifies player if needed. */
  private addMessage(resp: ServerMessage<{}>) {
    const sender = resp.player;
    const msg = resp.msg;

    const date = new Date();
    let hours = String(date.getHours());
    hours = ('00' + hours).substring(hours.length);
    let mins = String(date.getMinutes());
    mins = ('00' + mins).substring(mins.length);
    const time = `${hours}:${mins}`;
    const color = getColorForPlayer(sender);

    this.messages.push({
      color,
      sender: sender === this.playerID ? 'You' : sender,
      content: msg,
      time,
    });

    if (sender !== this.playerID && (!this.hasFocus || !this.drawerOpen)) {
      this.unreadMessages += 1;
    }

    if (!this.hasFocus && this.canNotify && sender !== this.playerID) {
      const n = new Notification(`Message from ${sender} in room ${this.roomJoined}`, {
        body: msg,
      });
    }

    // Offer gracious amount of time to render.
    setTimeout(() => {
      let el: any;
      if (this.$refs.lastMessage instanceof Array && this.$refs.lastMessage.length > 0) {
        el = this.$refs.lastMessage[0];
      }

      if (el.$el !== undefined) {
        el = el.$el;
      }

      el.scrollIntoView();
    }, 500);
  }

  /** We've been notified that some player has ditched all their cards. */
  private playerWon(resp: ServerMessage<{}>) {
    const idx = this.table.findIndex((i) => i.id === resp.player);
    this.winners[resp.player] = true;
    if (this.playerID === resp.player) {
      this.overlayMsgs.push(`Congrats! You've escaped!`);
    } else {
      this.overlayMsgs.push(`${resp.player} escapes.`);
    }
  }

  /** We've been notified that the game has ended. */
  private gameEnded(resp: ServerMessage<{}>) {
    if (this.playerID === resp.player) {
      this.overlayMsgs.push(`You've got leftover cards. You lose.`);
    } else {
      this.overlayMsgs.push(`${resp.player} has leftover card(s) and loses.`);
    }
  }

  private requestGameRestart() {
    this.restartRequest = false;
    this.conn.requestNewGmae(this.playerID, this.roomJoined!);
  }

  /** Sets the snackbar message. */
  private showError(msg: string) {
    this.notification = msg;
  }

  /* Helper methods */

  /** Shows alert as an alert notification in the app bar. */
  private showAlert(msg: string, ty: string = 'success') {
    this.alertType = ty;
    this.alertMsg = msg;
    setTimeout(() => {
      this.alertMsg = null;
    }, 3000);
  }

  /**
   * Initiates a cooldown time by locking the table and hence preventing
   * the players from placing any more cards. This is done in the end of
   * each round.
   *
   * @returns Timeout (in ms) until the table is locked.
   */
  private initiateTableLockdown(): number {
    const remaining = Math.floor(TABLE_LOCK_TIME_MS / 1000);
    this.tableLockTime = 100;
    for (let i = 1; i <= remaining; i++) {
      setTimeout(() => {
        this.tableLockTime = ((remaining - i) / remaining) * 100;
        if (i === remaining) {
          setTimeout(() => {
            this.tableLockTime = null;
          }, 500);
        }
      }, i * 1000);
    }

    return remaining * 1000;
  }

  /**
   * Updates the player's hand with new cards from the server.
   *
   * Even though we could maintain the same structure (`[]Card`) and let Vue
   * show the transitions, we don't, because we need to group based on suites
   * and sort them. This means we're (mostly) on our own. So, we find the diff
   * off the existing hand, and then go about adding/subtracting the cards.
   *
   * @param newHand Updated hand from the server.
   * @param timeout Timeout for each card update (add/remove).
   * @param initial Internal param for recursion.
   */
  private updateHand(newHand: Card[], timeout: number | null = null, initial: boolean = true) {
    if (initial) {
      // Mark all cards as old.
      this.hand.forEach((item) => {
        item.labels.forEach((l) => {
          l.updated = false;
        });
      });
    }

    setTimeout(() => {
      let card = newHand.pop();
      while (card !== undefined) {
        const i = suiteIndices[card.suite];
        const labels = this.hand[i].labels;
        const label = { label: card.label, updated: true };
        const j = searchSortedIndex(labels, label, (c1, c2) => labelRanks[c1.label] - labelRanks[c2.label]);
        if (labels[j] && labels[j].label === card.label) {
          // If this card already exists, then mark it as updated and progress.
          labels[j].updated = true;
          card = newHand.pop();
        } else {
          // This is a new card (+ diff). Recurse with a timeout.
          labels.splice(j, 0, label);
          this.updateHand(newHand, null, false);
          break;
        }
      }

      // If we don't have any cards left, then progress to removal.
      if (card === undefined) {
        let itemsEmpty = true;
        this.hand.forEach((item) => {
          const idx = item.labels.findIndex((c) => !c.updated);
          if (idx >= 0) {
            item.labels.splice(idx, 1);
          }

          itemsEmpty = itemsEmpty && idx < 0;
        });

        if (!itemsEmpty) {
          // We have removed cards in this run. Recurse (- diff) with a timeout.
          this.updateHand(newHand, null, false);
        }
      }
    }, timeout ? timeout : 50);
  }

  /** Player has requested for a tutorial. */
  private beginTutorial() {
    this.showTutorial = true;
    this.conn = new TutorialProvider();
    this.addListeners();
  }

  /** Player has progressed (to/fro) in tutorial. */
  private stepTutorial(step: TutorialStep) {
    if (step.room !== undefined) {
      this.roomJoined = step.room;
    }
  }

  /** Player has ended the tutorial. */
  private endTutorial() {
    this.initialize();
    // NOTE: We shouldn't add listeners again here (ConnectionProvider uses statics).
  }
}
export { ALLOWED_PLAYERS };
</script>

<style scoped>
* {
  transition: all 500ms ease;
}

.card-group-enter, .card-group-leave-to, .fade-enter, .fade-leave-to {
  opacity: 0;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity .5s;
}

.card-group-enter-active, .card-group-leave-active, .card-group-move {
  transition: all 1s;
}

</style>
