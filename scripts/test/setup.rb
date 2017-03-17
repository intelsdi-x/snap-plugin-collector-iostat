its_container(:snap) do
  describe 'setup iostat package' do
    it do
      c = cmd_with_retry('apt-get update && apt install -y sysstat')
      expect(c.exit_status).to eq 0
    end
  end
end
