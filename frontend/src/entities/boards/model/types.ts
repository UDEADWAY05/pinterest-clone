import { Pin } from '@/entities/pin/model/types';

export interface Board {
  id: number;
  title: string;
  pins: Pin[];
  userId: number;
}

export interface BoardWithRecentPins extends Board {
  recent_pins: Pin[];
}
