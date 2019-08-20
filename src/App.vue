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
        <v-divider></v-divider>
        <v-overlay opacity="0.5" :value="overlayTip !== null">
          <span>{{ overlayTip }}</span>
          <v-btn icon @click="overlayTip = null">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-overlay>
        <v-row class="d-flex justify-center mt-5">
          <v-tooltip right>
            <template v-slot:activator="{ on }">
              <v-btn icon :disabled="cardIndex === null" v-on="on" @click="sendToPile">
                <v-icon x-large>mdi-chevron-up-circle</v-icon>
              </v-btn>
            </template>
            <span>Add your card to the pile</span>
          </v-tooltip>
        </v-row>
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
      </v-container>
      <JoinRoom @joined="playerJoined" :players="allowedPlayers" :showDialog="!roomJoined" />
    </v-content>
    <v-snackbar :value="true" :timeout="0" v-for="(text, i) in notifications" :key="i">
      {{ text }}
      <v-btn color="pink" text @click="notifications.splice(i, 1)">Close</v-btn>
    </v-snackbar>
    <v-footer app></v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import JoinRoom from './dialog/JoinRoom.vue';
import { Card, Suite, suitePrettyMap, Label, PlayerCard } from './persistence/model';
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

  private roomJoined: boolean = true;

  private notifications: string[] = [];

  private alertMsg: string | null = null;

  private alertType: string = 'info';

  private overlayTip: string | null = null;

  private cardIndex: number | null = null;

  private hand: Card[] = [{
      label: Label.Six,
      suite: Suite.Clover,
    }, {
      label: Label.Six,
      suite: Suite.Clover,
    }];

  private table: PlayerCard[] = [{
    card: {
      label: Label.King,
      suite: Suite.Diamond,
    },
    id: 'player-1',
  }, {
    card: {
      label: Label.Six,
      suite: Suite.Clover,
    },
    id: 'player-2',
  }, {
    card: {
      label: Label.Queen,
      suite: Suite.Spade,
    },
    id: 'player-3',
  }, {
    card: {
      label: Label.Ace,
      suite: Suite.Diamond,
    },
    id: 'player-4',
  }, {
    card: {
      label: Label.Two,
      suite: Suite.Heart,
    },
    id: 'player-5',
  }, {
    card: {
      label: Label.Seven,
      suite: Suite.Clover,
    },
    id: 'player-6',
  }];

  /* Internal properties */

  private conn = new ConnectionProvider();

  private created() {
    this.conn.onError(this.showError, true);
    this.conn.onPlayerTurn((resp) => {
      this.hand = resp.response.hand;
      this.table = resp.response.table;
    }, true);
  }

  private playerJoined(self: string, resp: ServerMessage<RoomResponse>) {
    if (resp.player === self) {
      this.roomJoined = true;
    }

    const diff = resp.response.max - resp.response.players.length;
    if (diff > 0) {
      this.alertMsg = `Waiting for ${diff} more player(s).`;
    } else {
      this.alertMsg = 'Yay!';
      this.alertType = 'success';
      setTimeout(() => {
        this.alertMsg = null;
      }, 3000);
    }
  }

  private sendToPile() {
    //
  }

  private showError(msg: string) {
    this.notifications.push(msg);
  }
}
export { ALLOWED_PLAYERS };
</script>
