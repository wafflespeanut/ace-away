<template>
  <v-app id="inspire">
    <v-app-bar app dense>
      <v-toolbar-title></v-toolbar-title>
      <v-row class="d-flex justify-center">
        <v-slide-y-transition>
          <v-alert v-if="alertMsg"
                    border="top" colored-border
                    :type="alertType" elevation="2">{{ alertMsg }}</v-alert>
        </v-slide-y-transition>
      </v-row>
      <v-btn icon @click="showTutorial = true">
        <v-icon>mdi-help-circle-outline</v-icon>
      </v-btn>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-row class="my-5"></v-row>
        <v-row class="d-flex flex-column flex-md-row">
          <v-row v-if="table.length" :class="{
            'mx-10': $vuetify.breakpoint.mdAndUp,
            'mb-10': true,
          }">
            <v-card :width="tableSize" :height="tableSize" class="mx-auto my-auto"
                    :style="{ borderRadius: '50%' }">
              <transition-group name="fade">
                <v-card :style="tableCardStyles(i)"
                        class="d-flex flex-column align-center justify-center"
                        v-for="(item, i) in table" :key="i + 0"
                        :height="0.9 * cardSize" :width="0.9 * cardSize"
                        :elevation="item.turn ? 8 : 2"
                        :color="item.won ? 'green darken-3' : (item.turn ? 'cyan darken-3' : '' )">
                  <div v-if="item.card" class="headline">{{ item.card.label }} {{ prettyMap[item.card.suite] }}</div>
                  <div v-else-if="item.turn">???</div>
                  <div v-else-if="item.won">[no cards]</div>
                  <div v-else>-</div>
                  <div class="caption">{{ item.id }}</div>
                </v-card>
              </transition-group>
            </v-card>
          </v-row>
          <v-row :style="{
            'flex-grow': $vuetify.breakpoint.mdAndUp ? '0' : '1',
          }" class="d-flex justify-center align-center mt-5 mb-5">
            <v-tooltip v-if="roomJoined" bottom>
              <template v-slot:activator="{ on }">
                <v-fab-transition>
                  <v-btn fab icon
                         :width="$vuetify.breakpoint.xs ? 80 : 100"
                         :height="$vuetify.breakpoint.xs ? 80 : 100"
                         :color="selectedCard ? (selectedCard.suite === 'h' || selectedCard.suite === 'd') ? 'red' : 'blue' : ''"
                         :disabled="cardIndex === null || tableLocked"
                         v-on="$vuetify.breakpoint.smAndUp ? on : () => {}"
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
          <v-col :cols="$vuetify.breakpoint.mdAndUp ? '5' : 'auto'" :class="{
            'mx-5': $vuetify.breakpoint.xs,
            'mx-10': $vuetify.breakpoint.smAndUp,
            'my-auto': true,
          }">
            <!-- We're "conditionally mandating" because we don't need a card selected all the time. -->
            <v-item-group :mandatory="cardIndex !== null" v-model="cardIndex">
              <v-row class="d-flex my-4" v-for="(item, x) in hand" :key="x">
                  <transition-group class="d-flex" name="glow">
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
        <v-overlay opacity="0.5" v-if="overlayMsg !== null">
          <div>
            <span class="display-1">{{ overlayMsg }}</span>
            <v-btn class="mb-4 ml-3" icon color="blue" @click="overlayMsg = null">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </div>
        </v-overlay>
      </v-container>
      <JoinRoom @player-set="v => playerID = v" :players="allowedPlayers" :showDialog="roomJoined === null" />
      <Tutorial @close="showTutorial = false" @tutorial-step="stepTutorial" :showDialog="showTutorial" />
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
import Tutorial from './dialog/Tutorial.vue';
import {
  Card, Suite, suitePrettyMap, Label, PlayerCard,
  GameEvent, suiteIndices, labelRanks,
} from './persistence/model';
import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './persistence/model';
import ConnectionProvider from './persistence/connection';
import GameEventHub from './persistence';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];
const START_ANGLE = Math.PI / 2;
const TABLE_LOCK_TIME = 5000;

const iconMap = {
  h: 'mdi-cards-heart',
  s: 'mdi-cards-spade',
  c: 'mdi-cards-club',
  d: 'mdi-cards-diamond',
};

interface CardLabel {
  label: string;
  updated: boolean;
}

interface HandItem {
  suite: string;
  labels: CardLabel[];
}

interface TableItem {
  id: string;
  card: Card | null;
  turn: boolean;
  won: boolean;
}

function searchSortedIndex<T>(items: T[], newItem: T, compare: (e1: T, e2: T) => number) {
  if (items.length === 0 || compare(items[0], newItem) > 0) {
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

@Component({
  components: {
    JoinRoom,
    Tutorial,
  },
})
export default class App extends Vue {

  /* Constants used by models */

  /** Allowed choices for players in rooms. */
  private allowedPlayers: number[] = ALLOWED_PLAYERS;

  /** Object for mapping suites to their unicode representations. */
  private prettyMap: any = suitePrettyMap;

  /** Object for mapping suites to their indices in hand. */
  private indexMap: any = suiteIndices;

  /** Object for mapping suites to their MD icons. */
  private iconMap: any = iconMap;

  /* Models */

  private showTutorial: boolean = false;

  /** Name set by the player (after creating/joining a room). */
  private playerID: string = '';

  /** Name of the room joined by the player (set after creating/joining a room). */
  private roomJoined: string | null = null;

  /** Notifications from the server shown as a snackbar at the bottom. */
  private notification: string | null = null;

  /** Alert message shown in app bar. */
  private alertMsg: string | null = null;

  /** Type of the alert. */
  private alertType: string = 'info';

  /** Message shown in overlay. */
  private overlayMsg: string | null = null;

  /** Card selected by the player. */
  private selectedCard: Card | null = null;

  /** Index of `selectedCard`. **Only used for resetting selection. Don't use this directly.** */
  private cardIndex: number | null = null;

  /** Player's hand containing their cards sorted by their labels and suites. */
  private hand: HandItem[] = [];

  /** Table containing the cards from all players for that round. */
  private table: TableItem[] = [];

  private previousTurnLength: number = 0;

  private tableLocked: boolean = false;

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

  /* Internal properties */

  private conn: GameEventHub = new ConnectionProvider();

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

  private created() {
    this.hand = Object.keys(suiteIndices)
      .sort((a, b) => suiteIndices[a] - suiteIndices[b])
      .map((s) => {
        return {
          suite: s,
          labels: [],
        };
      });

    this.conn.onError(this.showError, true);

    this.conn.onPlayerJoin((resp) => {
      // Set player states.
      this.table = resp.response.players.map((id, i) => {
        return {
          id,
          card: null,
          turn: resp.response.turnIdx === i,
          won: resp.response.escaped.indexOf(id) >= 0,
        };
      });

      if (resp.player === this.playerID) {
        this.roomJoined = resp.room;
      }

      // Notify if we're waiting on player(s) joining.
      const diff = resp.response.max - resp.response.players.length;
      if (diff > 0) {
        this.alertMsg = `Waiting for ${diff} more player(s) in room ${resp.room}.`;
      } else {
        this.alertMsg = `Yay! Let's begin!`;
        this.alertType = 'success';
        setTimeout(() => {
          this.alertMsg = null;
        }, 3000);
      }
    }, true);

    this.conn.onPlayerTurn((resp) => {
      const updateStuff = () => {
        // Sort the hand based on suites followed by labels.
        this.updateHand(resp.response.hand);
        this.previousTurnLength = resp.response.table.length;

        // Reset states of cards in our table (if the table isn't locked).
        this.table.forEach((v) => {
          v.won = false;
          v.card = null;
          v.turn = v.id === resp.response.turnPlayer;
        });

        // Get the cards and set them in our table.
        resp.response.table.forEach((c) => {
          const idx = this.table.findIndex((v) => v.id === c.id); // This will exist.
          this.table[idx].card = c.card;
        });
      };

      if (resp.response.table.length < this.previousTurnLength) {
        // If the table is getting cleared, lock the table and
        // pause for a moment for users to see what happened.
        // We're fine delaying this turn because we don't allow the
        // users to place a card in this interval, and so we won't
        // get any `playerTurn` events.
        this.tableLocked = true;
        setTimeout(() => {
          this.alertType = 'success';
          this.alertMsg = 'Table cleared!';
          setTimeout(() => {
            this.alertMsg = null;
          }, 3000);

          updateStuff();
          this.tableLocked = false;
        }, TABLE_LOCK_TIME);
      } else {
        updateStuff();
      }
    }, true);

    this.conn.onPlayerWin((resp) => {
      const idx = this.table.findIndex((i) => i.id === resp.player);
      this.table[idx].won = true;
      if (this.playerID === resp.player) {
        this.overlayMsg = `Congrats! You've escaped!`;
      } else {
        this.overlayMsg = `${resp.player} escapes.`;
      }
    });

    this.conn.onGameOver((resp) => {
      if (this.playerID === resp.player) {
        this.overlayMsg = `You've got leftover cards. You lose.`;
      } else {
        this.overlayMsg = `${resp.player} has leftover card(s) and loses.`;
      }
    });
  }

  private updateHand(newHand: Card[], timeout: number | null = null, initial: boolean = true) {
    if (initial) {
      this.hand.forEach((item) => {
        item.labels.forEach((l) => {
          l.updated = false;
        });
      });
    }

    setTimeout(() => {
      const card = newHand.pop();
      if (card === undefined) {
        let noMoreItems = true;
        this.hand.forEach((item) => {
          const idx = item.labels.findIndex((c) => !c.updated);
          if (idx >= 0) {
            item.labels.splice(idx, 1);
          }

          noMoreItems = noMoreItems && idx < 0;
        });

        if (!noMoreItems) {
          this.updateHand(newHand, null, false);
        }

        return;
      }

      const i = suiteIndices[card.suite];
      const labels = this.hand[i].labels;
      const label = { label: card.label, updated: true };
      const j = searchSortedIndex(labels, label, (c1, c2) => labelRanks[c1.label] - labelRanks[c2.label]);
      const delCount = labels[j] && labels[j].label === card.label ? 1 : 0;
      labels.splice(j, delCount, label);
      this.updateHand(newHand, null, false);
    }, timeout ? timeout : 50);
  }

  /** Sends the player-selected card to the pile of cards in the table. */
  private sendToPile() {
    console.log(`Player placing ${this.selectedCard!.label}${this.prettyMap[this.selectedCard!.suite]}`);
    this.conn.showCard({
      player: this.playerID,
      room: this.roomJoined!,
      event: GameEvent.playerTurn,
      data: {
        card: this.selectedCard!,
      },
    });

    this.cardIndex = null;
  }

  private stepTutorial() {
    //
  }

  /** Sets the snackbar message. */
  private showError(msg: string) {
    this.notification = msg;
  }
}
export { ALLOWED_PLAYERS };
</script>

<style scoped>
* {
  transition: all 500ms ease;
}

.glow-enter, .glow-leave-to, .fade-enter, .fade-leave-to {
  opacity: 0;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity .5s;
}

.glow-enter-active, .glow-leave-active {
  animation: glow 1s;
  box-shadow: 0 0 0 2em rgba(255, 255, 255, 0);
}

.glow-enter-active, .glow-leave-active, .glow-move {
  transition: all 1s;
}

@keyframes glow {
  0% {
    box-shadow: 0 0 0 0 #ff8a65;
  }
}
</style>
