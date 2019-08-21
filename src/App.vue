<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawerOpen" app>
      <!--  -->
    </v-navigation-drawer>

    <v-app-bar app>
      <v-app-bar-nav-icon @click.stop="drawerOpen = !drawerOpen"></v-app-bar-nav-icon>
      <v-toolbar-title>Ace Away!</v-toolbar-title>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-slide-y-transition>
          <v-alert v-if="alertMsg" border="top" colored-border :type="alertType" elevation="2">{{ alertMsg }}</v-alert>
        </v-slide-y-transition>
        <!-- We're gating "mandatory" because we don't need a card selected by default. -->
        <v-item-group :mandatory="cardIndex !== null" v-model="cardIndex">
          <v-row>
            <v-col class="d-flex justify-center" v-for="(card, i) in hand" :key="i">
              <v-item v-slot:default="{ active, toggle }">
                <v-slide-x-transition>
                  <v-card :color="active ? ( card.suite == 'h' || card.suite == 'd' ? 'red' : 'blue' ) : ''"
                          class="d-flex align-center justify-center"
                          height="80"
                          width="80"
                          @click="toggle">
                    <span class="display-1" :class="active ? 'flex-grow-1 text-center' : ''">{{ card.label }} {{ prettyMap[card.suite] }}</span>
                  </v-card>
                </v-slide-x-transition>
              </v-item>
            </v-col>
          </v-row>
        </v-item-group>
        <v-row class="d-flex justify-center mt-5">
          <v-tooltip right>
            <template v-slot:activator="{ on }">
              <v-fab-transition>
                <v-btn icon
                       fab
                       color="pink"
                       :disabled="cardIndex === null"
                       v-on="on"
                       @click="sendToPile">
                  <v-icon x-large>mdi-chevron-down-circle</v-icon>
                </v-btn>
              </v-fab-transition>
            </template>
            <span>Add your card to the pile</span>
          </v-tooltip>
        </v-row>
        <v-divider></v-divider>
        <v-row>
          <v-col class="d-flex justify-center" v-for="(item, i) in table" :key="i">
            <v-slide-x-transition>
              <v-card class="d-flex flex-column align-center justify-center"
                      height="100"
                      width="100">
                <div class="display-1">{{ item.card.label }} {{ prettyMap[item.card.suite] }}</div>
                <div class="caption">{{ item.id }}</div>
              </v-card>
            </v-slide-x-transition>
          </v-col>
        </v-row>
        <v-overlay opacity="0.5" v-if="overlayTip !== null">
          <div>
            <span class="display-1">{{ overlayTip }}</span>
            <v-btn class="mb-4 ml-3" icon color="blue" @click="overlayTip = null">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </div>
        </v-overlay>
        <v-row class="d-flex justify-center" v-for="(msg, i) in temporaryMessages" :key="i">
          <v-slide-x-transition>
            <span class="headline mt-3">{{ msg }}</span>
          </v-slide-x-transition>
        </v-row>
      </v-container>
      <JoinRoom @joined="playersChanged" :players="allowedPlayers" :showDialog="roomJoined === null" />
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
import { Card, Suite, suitePrettyMap, Label, PlayerCard, GameEvent } from './persistence/model';
import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './persistence/model';
import ConnectionProvider from './persistence/connection';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];

@Component({
  components: {
    JoinRoom,
  },
})
export default class App extends Vue {

  /* Constants used by models */

  private allowedPlayers: number[] = ALLOWED_PLAYERS;

  private prettyMap: any = suitePrettyMap;

  /* Models */

  private drawerOpen: boolean = false;

  private roomJoined: string | null = null;

  private notification: string | null = null;

  private alertMsg: string | null = null;

  private alertType: string = 'info';

  private overlayTip: string | null = null;

  private cardIndex: number | null = null;

  private hand: Card[] = [];

  private table: PlayerCard[] = [];

  private playerID: string = '';

  private temporaryMessages: string[] = [];

  /* Internal properties */

  private conn = new ConnectionProvider();

  private created() {
    this.conn.onError(this.showError, true);
    this.conn.onPlayerTurn((resp) => {
      this.hand = resp.response.hand;
      this.table = resp.response.table;

      if (resp.response.table.length === 0) {
        this.postMessage('Pile has been cleared!');
      }

      if (resp.response.ourTurn) {
        this.postMessage('Your turn.', 4);
      } else {
        this.postMessage(`It's ${resp.response.turnPlayer}'s turn.`, 3);
      }
    }, true);

    this.conn.onPlayerWin((resp) => {
      this.overlayTip = `${resp.player} heroically leaves the room.`;
    });

    this.conn.onGameOver((resp) => {
      this.overlayTip = `${resp.player} has leftover card(s) and loses.`;
    });
  }

  private playersChanged(self: string, resp: ServerMessage<RoomResponse>) {
    this.playerID = self;
    if (resp.player === self) {
      this.roomJoined = resp.room;
    }

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
  }

  private sendToPile() {
    this.conn.showCard({
      player: this.playerID!,
      room: this.roomJoined!,
      event: GameEvent.playerTurn,
      data: {
        card: this.hand[this.cardIndex!],
      },
    });

    this.cardIndex = null;
  }

  private postMessage(msg: string, timeout?: number) {
    this.temporaryMessages.push(msg);
    setTimeout(() => {
      this.temporaryMessages.splice(this.temporaryMessages.indexOf(msg), 1);
    }, (timeout == null) ? 2000 : timeout * 1000);
  }

  private showError(msg: string) {
    this.notification = msg;
  }
}
export { ALLOWED_PLAYERS };
</script>
