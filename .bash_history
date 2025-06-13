apt update
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
apt install docker-compose-plugin -y
ls
cd ..
ls
mkdir -p /root/dutch-bot
ls
cd root
ls
cd dutch-bot/
ls
cat .env
./deploy.sh 
cd ..
docker
docker compose
cd -
ls
nano deploy.sh 
./deploy.sh 
nano Dockerfile 
./deploy.sh 
ls
cat Dockerfile 
./deploy.sh 
cat Dockerfile 
./deploy.sh 
./run_bot.sh 
apt install go
snap install go
go version
apt get go
apt install go
apt install golang-go
go version
./run_bot.sh 
ls
nohup ./bot > bot.log 2>&1 &
tail -f boy.log
tail -f bot.log
ls
cat bot.log 
nohup ./bot > bot.log 2>&1 &
cat bot.log 
ls
./run_bot.sh 
nohup ./bot > bot.log 2>&1 &
cat bot.log 
rm bot.log
nohup ./bot > bot.log 2>&1 &
cat bot.log 
cat .env
./bot
ls
tail -f bot.log
./bot
./fix-architecture.sh
tail -f bot.log
cd ..
fallocate -l 1G /swapfile
chmod 600 /swapfile 
mkdswap /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab
free -h
ls
cd dutch-bot/
ls
./deploy.sh 
docker ps
docker compose ps
docker compose logs -f dutch-bot-dutch-bot-1
docker compose ps
docker compose logs -f dutch-bot
./deploy.sh 
ls
cd ..
ls
cd -
ls
./deploy.sh 
ls
./deploy.sh 
docker compose ps
docker compose logs -f dutch-bot-dutch-bot-1
docker compose ps
docker compose logs -f dutch-bot
./deploy.sh 
docker-compose logs -f
docker compose logs -f
docker compose ps
ls
./bot 
chmod +x bot
./bot 
./run_bot.sh 
docker compose down
./quick-deploy.sh
ls
cat bot.log 
cd ..
ls
./bot
cd ..
ls
cd root/
ls
ls -la
rm -rf *
ls -la
cd -
pwd
cd root/
pwd
ls
chmod +x deploy.sh
./deploy.sh 
docker compose logs -f
ls
docker system prune -af --volumes
ls
./run_bot.sh 
screen -S dutch-bot
screen -l
screen
screen -r
./deploy.sh 
./deploy_with_migration.sh 
pkill -f "dutch-bot" || true
./run_bot.sh 
screen -l dutch-bot
screen -l
sqlite3 dutch_learning.db "
SELECT u.telegram_id, u.first_name, u.last_active, COUNT(up.id) as progress_count
FROM users u 
LEFT JOIN user_progress up ON u.id = up.user_id 
GROUP BY u.id;"
sqlite3 dutch_learning.db "
SELECT u.telegram_id, COUNT(up.id) as due_words
FROM users u 
JOIN user_progress up ON u.id = up.user_id 
WHERE up.due_date <= datetime('now')
GROUP BY u.id;"
ls
screen -l
cp dutch_bot.db dutch_bot.db.backup.manual
sqlite3 dutch_bot.db < migrate_categories.sql
ls
pkill -f "bot" || true
cp dutch_learning.db dutch_learning.db.backup.$(date +%Y%m%d_%H%M%S)
sqlite3 dutch_learning.db "SELECT category, COUNT(*) FROM words GROUP BY category ORDER BY COUNT(*) DESC;"
sqlite3 dutch_learning.db < migrate_categories.sql
sqlite3 dutch_learning.db "SELECT category, COUNT(*) FROM words GROUP BY category ORDER BY COUNT(*) DESC;"
screen -dmS dutch-bot bash -c './run_bot.sh 2>&1 | tee -a logs/bot_full.log'
screen -r dutch-bot
ls
ls -la
cd logs/
ls
cat bot_full.log 
screen -r
screen -r dutch-bot
cd -
ls
rm dutch_bot.db dutch_bot.db.backup.20250611_112427 dutch_bot.db.backup.20250611_112549 dutch_bot.db.backup.manual
ls
rm deploy-w
rm deploy_with_migration.sh 
ls
sqlite3 dutch_learning.db "SELECT category, COUNT(*) FROM grammar_tips GROUP BY category;"
pkill -f "bot" || pkill -f "dutch"
ps aux | grep bot
screen -dmS dutch-bot bash -c './run_bot.sh 2>&1 | tee -a logs/bot_full.log'
screen -r dutch-bot
sqlite3 dutch_learning.db "SELECT COUNT(*) as total_tips FROM grammar_tips;"
cat grammar_tips.json | jq '.grammar_tips[] | select(.applicable_categories[] | contains("road_signs")) | .title'
cat grammar_tips.json | jq -r '.grammar_tips[].applicable_categories[]' | sort | uniq -c
screen -r dutch-bot
ls
cd internal/
ls
cd ../logs/
ls
cat bot_full.log 
