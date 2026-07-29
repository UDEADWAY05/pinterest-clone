export interface CreatePinRequest {
  title: string;
  description: string;
  image_url: string;
  width: number;
  height: number;
  dominant_color: string;
  tags?: string[];
}

export interface PresignRequest {
  file_name: string;
  content_type: string;
  file_size: number;
}

export interface PresignResult {
  upload_url: string;
  public_url: string;
  key: string;
  expires_in: number;
}
