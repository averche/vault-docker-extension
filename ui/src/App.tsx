import React from 'react';
import Button from '@mui/material/Button';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import { Stack, TextField, Typography } from '@mui/material';

// Note: This line relies on Docker Desktop's presence as a host application.
// If you're running this React app in a browser, it won't work properly.
const client = createDockerDesktopClient();

function useDockerDesktopClient() {
  return client;
}

export function App() {
  const [secrets, setSecrets] = React.useState<string>();

  const [readKey, setReadKey] = React.useState<string>("sample-secret1");
  const [readSecret, setReadSecret] = React.useState<string>();

  const [writeKey, setWriteKey] = React.useState<string>("sample-secret1");
  const [writeSecret, setWriteSecret] = React.useState<string>();

  const ddClient = useDockerDesktopClient();

  const fetchSecrets = async () => {
    const result = await ddClient.extension.vm?.service?.get('/get-secrets');
    setSecrets(Array.prototype.slice.call(result).join("\n"));
  };

  const fetchSecret = async () => {
    const result = await ddClient.extension.vm?.service?.get('/get-secret?key='+readKey);
    setReadSecret(JSON.stringify(result));
  };

  const pushSecret = async () => {
    const result = await ddClient.extension.vm?.service?.post(
      '/write-secret?key='+writeKey,
      JSON.parse(writeSecret)
    );
  };

  return (
    <>
      <Typography variant="h1">HashiCorp Vault</Typography>
      <Typography variant="h3" sx={{ mt: 2 }} >Access</Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
        Use the following address and token to access this vault instance:
      </Typography>
      <Typography variant="body1" color="text.secondary" fontFamily={"Courier"} sx={{ mt: 1 }}>
        export VAULT_ADDR=http://127.0.0.1:8201<br/>
        export VAULT_TOKEN=root
      </Typography>
      <Typography variant="h3" sx={{ mt: 3 }} >Read Secret</Typography>
      <Stack direction="row" alignItems="start" spacing={2} sx={{ mt: 1 }}>
        <Stack direction="column" alignItems="start" spacing={2} >
          <TextField
            label="Key"
            sx={{ width: 160 }}
            variant="outlined"
            id="my-field"
            defaultValue="sample-secret1"
            onChange={e => setReadKey(e.target.value)}
          />
          <Button variant="contained" sx={{ width: 160 }} onClick={fetchSecret}>
           Read
          </Button>
        </Stack>
        <TextField
          label="Secret"
          sx={{ width: 480 }}
          disabled
          multiline
          variant="outlined"
          minRows={4}
          value={readSecret ?? ''}
        />
      </Stack>
      <Typography variant="h3" sx={{ mt: 3 }} >Write Secret</Typography>
      <Stack direction="row" alignItems="start" spacing={2} sx={{ mt: 1 }}>
        <Stack direction="column" alignItems="start" spacing={2} >
          <TextField
            label="Key"
            sx={{ width: 160 }}
            variant="outlined"
            id="my-write-field"
            defaultValue="sample-secret1"
            onChange={e => setWriteKey(e.target.value)}
          />
          <Button variant="contained" sx={{ width: 160 }} onClick={pushSecret}>
           Write
          </Button>
        </Stack>
        <TextField
          label="Secret"
          sx={{ width: 480 }}
          multiline
          variant="outlined"
          minRows={4}
          onChange={e => setWriteSecret(e.target.value)}
        />
      </Stack>
      <Typography sx={{ mt: 3 }} variant="h3">Secrets</Typography>
      <Stack direction="row" alignItems="start" spacing={2} sx={{ mt: 1 }}>
        <Button variant="contained" sx={{ width: 160 }} onClick={fetchSecrets} onLoad={fetchSecrets}>
          Refresh
        </Button>
        <TextField
          label="Secrets"
          sx={{ width: 480 }}
          disabled
          multiline
          variant="outlined"
          minRows={4}
          value={secrets ?? ''}
        />
      </Stack>
    </>
  );
}
