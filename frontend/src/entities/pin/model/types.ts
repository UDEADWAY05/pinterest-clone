export interface PinAuthor {
  id: number;
  display_name: string;
  avatar_url: string;
}

export interface Pin {
  id: number;
  title: string;
  image_url: string;
  description: string;
  width: number;
  height: number;
  dominant_color: string;
  author?: PinAuthor;
  tags: string[];
  saves_count: number;
  is_saved: boolean;
  likes_count: number;
}
