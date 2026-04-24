import { api } from '@/api/client';

export type FileCategory = 'sites' | 'profiles' | 'users';

export type ManagedFile = {
  id: string;
  name: string;
  size_bytes: number;
  category: FileCategory;
  uploaded_at?: string;
  download_url?: string;
};

export const getFiles = (category: FileCategory) =>
  api
    .get<ManagedFile[]>('/files', { params: { category } })
    .then(({ data }) => data);

export const uploadFile = (
  file: File,
  category: FileCategory,
  onProgress?: (progress: number) => void
) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('category', category);

  return api
    .post<ManagedFile>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (event) => {
        if (!event.total || !onProgress) {
          return;
        }
        onProgress(Math.round((event.loaded / event.total) * 100));
      },
      maxBodyLength: Infinity,
      maxContentLength: Infinity
    })
    .then(({ data }) => data);
};
