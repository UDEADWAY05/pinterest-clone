export interface User {
  id: number;
  email: string;
  display_name: string;
  avatar_url: string;
  created_at: number;
  bio: string;
  is_current_user: boolean;
}

export interface UserProfile extends User {
  pins_count: number;
  followers_count: number;
  following_count: number;
  is_following: boolean;
}
