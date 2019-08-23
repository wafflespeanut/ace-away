<template>
  <v-app id="inspire">
    <v-app-bar app>
      <v-row class="d-flex justify-center">
        <v-slide-y-transition>
          <v-alert v-if="alertMsg"
                    border="top" colored-border
                    :type="alertType" elevation="2">{{ alertMsg }}</v-alert>
        </v-slide-y-transition>
      </v-row>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-row class="my-5"></v-row>
        <v-row class="d-flex flex-column flex-md-row">
          <v-row v-if="table.length" class="mx-8 mb-10">
            <v-card :width="tableSize" :height="tableSize" class="mx-auto my-auto"
                    :style="{ borderRadius: '50%' }">
              <v-card :style="tableCardStyles(i)"
                      class="d-flex flex-column align-center justify-center"
                      v-for="(item, i) in table" :key="i"
                      :height="0.9 * cardSize" :width="0.9 * cardSize"
                      :elevation="item.turn ? 8 : 2"
                      :color="item.won ? 'green darken-3' : (item.turn ? 'cyan darken-3' : '' )">
                <div v-if="item.won">[no cards]</div>
                <div v-else-if="item.turn">???</div>
                <div v-else-if="item.card" class="headline">{{ item.card.label }} {{ prettyMap[item.card.suite] }}</div>
                <div v-else>-</div>
                <div class="caption">{{ item.id }}</div>
              </v-card>
            </v-card>
          </v-row>
          <v-row class="d-flex justify-center align-center mt-5 mb-10">
            <v-tooltip bottom v-if="hand.length">
              <template v-slot:activator="{ on }">
                <v-fab-transition>
                  <v-btn icon
                        fab
                        color="red"
                        :disabled="cardIndex === null"
                        v-on="on"
                        @click="sendToPile">
                    <v-icon class='display-3'>mdi-fire</v-icon>
                  </v-btn>
                </v-fab-transition>
              </template>
              <span>Add your card to the pile</span>
            </v-tooltip>
          </v-row>
          <v-row class="mx-8 my-auto">
            <!-- We're gating "mandatory" because we don't need a card selected by default. -->
            <v-item-group :mandatory="cardIndex !== null" v-model="cardIndex">
              <v-row v-for="(icon, suite, x) in prettyMap" :key="x">
                <!-- Filtering here doesn't affect `cardIndex` because we've already sorted the hand. -->
                <v-col>
                  <v-row>
                    <v-item v-for="(card, i) in hand.filter((c) => c.suite === suite)" :key="i" v-slot:default="{ active, toggle }">
                      <v-card :color="active ? ( suite == 'h' || suite == 'd' ? 'red' : 'blue' ) : ''"
                              @click="toggle"
                              :height="cardSize"
                              :width="cardSize"
                              class="d-flex justify-center align-center">
                        <span class="display-1">{{ card.label }} {{ icon }}</span>
                      </v-card>
                    </v-item>
                  </v-row>
                </v-col>
              </v-row>
            </v-item-group>
          </v-row>
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
import { Card, Suite, suitePrettyMap, Label, PlayerCard, GameEvent, suiteRanks, labelRanks } from './persistence/model';
import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './persistence/model';
import ConnectionProvider from './persistence/connection';
import GameEventHub from './persistence';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];
const START_ANGLE = Math.PI / 2;

interface TableItem {
  id: string;
  card: Card | null;
  turn: boolean;
  won: boolean;
}

@Component({
  components: {
    JoinRoom,
  },
})
export default class App extends Vue {

  /* Constants used by models */

  /**
   * Allowed choices for players in rooms.
   */
  private allowedPlayers: number[] = ALLOWED_PLAYERS;

  /**
   * Object for mapping suites to their unicode representations.
   */
  private prettyMap: any = suitePrettyMap;

  /* Models */

  /**
   * Player ID set by the user (after creating/joining a room).
   */
  private playerID: string = '';

  /**
   * Name of the room joined by the user (set after creating/joining a room).
   */
  private roomJoined: string | null = null;

  /**
   * Notifications from the server shown as a snackbar at the bottom.
   */
  private notification: string | null = null;

  /**
   * Alert message shown in app bar.
   */
  private alertMsg: string | null = null;

  /**
   * Type of the alert.
   */
  private alertType: string = 'info';

  /**
   * Message shown in overlay.
   */
  private overlayMsg: string | null = null;

  /**
   * Index of the card selected by the user.
   */
  private cardIndex: number | null = null;

  /**
   * Player's hand containing their cards sorted by their labels and suites.
   */
  private hand: Card[] = [];

  /**
   * Table containing the cards from all players for that round.
   */
  private table: TableItem[] = [];

  /**
   * Size of the card based on viewports.
   */
  private get cardSize(): number {
    if (screen.width <= 600) {
      return 85;
    } else if (screen.width <= 960) {
      return 95;
    } else {
      return 100;
    }
  }

  /**
   * Size of the table based on viewports.
   */
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

  /**
   * Stypes applied to the cards within the table.
   */
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
      // Sort the hand based on suites followed by labels.
      this.hand = resp.response.hand.sort((c1, c2) => {
        return suiteRanks[c1.suite] * labelRanks[c1.label] - suiteRanks[c2.suite] * labelRanks[c2.label];
      });

      // Reset states of cards in our table.
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

  /**
   * Sends the user-selected card to the pile of cards in the table.
   */
  private sendToPile() {
    this.conn.showCard({
      player: this.playerID,
      room: this.roomJoined!,
      event: GameEvent.playerTurn,
      data: {
        card: this.hand[this.cardIndex!],
      },
    });

    this.cardIndex = null;
  }

  /**
   * Set snackbar message.
   */
  private showError(msg: string) {
    this.notification = msg;
  }
}
export { ALLOWED_PLAYERS };
</script>
