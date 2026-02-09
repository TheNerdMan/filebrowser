interface PopupProps {
  prompt: string;
  confirm?: any;
  action?: PopupAction;
  saveAction?: () => void;
  props?: any;
  close?: (() => Promise<string>) | null;
  path?: string;
}

type PopupAction = (e: Event) => void;
