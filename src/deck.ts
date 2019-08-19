
interface Card {
  label: Label;
  suite: Suite;
}

enum Suite {
  Diamond = 'd',
  Clover  = 'c',
  Heart   = 'h',
  Spade   = 's',
}

const suitePrettyMap: any = {
  d: '♦',
  c: '♣',
  h: '♥',
  s: '♠',
};

enum Label {
  Two   = '2',
  Three = '3',
  Four  = '4',
  Five  = '5',
  Six   = '6',
  Seven = '7',
  Eight = '8',
  Nine  = '9',
  Ten   = '10',
  Jack  = 'J',
  Queen = 'Q',
  King  = 'K',
  Ace   = 'A',
}

export { Card, Label, Suite, suitePrettyMap };
